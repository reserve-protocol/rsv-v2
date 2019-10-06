// Package soltools provides a Go-to-JavaScript bridge to use 0x's suite of sol-X tools from Go.
//
// 0x has a suite of tools for Solidity development:
//
//	https://sol-coverage.com/
//	https://sol-compiler.com/
//	https://sol-trace.com/
//	https://sol-profiler.com/
//
// The tools are designed to be used from JavaScript. This package provides a bridge so they can be used
// from Go. The primary interface is Backend, which is a replacement for an *ethclient.Client that sends
// transactions and calls through a 0x library that adds tracing.
package soltools

import (
	"bufio"
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
)

// Backend is a replacement for an *ethclient.Client that sends transactions through 0x's tracing
// library in JavaScript.
type Backend struct {
	*ethclient.Client
	cmd           *exec.Cmd
	waitForStdout sync.WaitGroup
}

// NewBackend dials an ethereum node at nodeAddress and returns a *Backend client for that node.
//
// NewBackend also starts a Node.js process, which the caller is responsible for closing by calling
// Backend.Close(). Example:
//
//	backend, err := NewBackend("http://localhost:8545", "project/artifacts", "project/contracts")
//	// handle err
//	defer backend.Close()
//
// The client will add tracing to the Ethereum transactions and calls that are made through it.
// It can also write a coverage report, which requires passing paths to artifacts and contracts
// directories for the corresponding Solidity code.
func NewBackend(nodeAddress string) (*Backend, error) {
	goBackend, err := ethclient.Dial(nodeAddress)
	if err != nil {
		return nil, err
	}

	repoDir := os.Getenv("REPO_DIR")
	if repoDir == "" {
		return nil, errors.New("REPO_DIR env var is not set -- need it to point to repo root")
	}
	bridgeJSPath := filepath.Join(repoDir, "soltools", "bridge.js")

	artifactsDir := filepath.Join(repoDir, "artifacts")
	contractsDir := filepath.Join(repoDir, "contracts")

	// check that the current working directory is where we think it is
	if _, err := os.Stat(bridgeJSPath); err != nil {
		if os.IsNotExist(err) {
			return nil, errors.Errorf("could not find %v -- make sure this command is running from the repo root", bridgeJSPath)
		}
		return nil, errors.Wrapf(err, "could not find %v", bridgeJSPath)
	}

	// copy to stdout and watch for starting line

	cmd := exec.Command("node", bridgeJSPath, artifactsDir, contractsDir)
	cmd.Stdin = os.Stdin
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	result := &Backend{
		Client: goBackend,
		cmd:    cmd,
	}

	bufferedStdout := bufio.NewReader(stdout)
	for {
		line, err := bufferedStdout.ReadString('\n')
		fmt.Println(strings.TrimSpace(line))
		if err != nil {
			result.Close()
			return nil, err
		}
		if strings.Contains(line, "server listening") {
			break
		}
	}

	result.waitForStdout.Add(1)
	go func() {
		defer result.waitForStdout.Done()
		io.Copy(os.Stdout, bufferedStdout)
	}()

	return result, nil
}

// Close frees resources associated with this Backend.
//
// In particular, it closes the backing JavaScript process and a network connection to the Ethereum node.
func (b *Backend) Close() error {
	b.Client.Close()
	err := b.call(
		"close",
		true,      // ignored input
		new(bool), // ignore output
	)
	if err != nil {
		b.cmd.Process.Kill()
		return err
	}
	b.waitForStdout.Wait()
	return b.cmd.Wait()
}

// call makes HTTP calls to the Node.js process.
func (*Backend) call(method string, in, out interface{}) error {
	b, err := json.Marshal(map[string]interface{}{
		"method": method,
		"data":   in,
	})
	if err != nil {
		return err
	}
	resp, err := http.Post("http://localhost:3000", "application/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case 200:
		return json.NewDecoder(resp.Body).Decode(out)
	case 500:
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("%s", b)
	default:
		return fmt.Errorf("unexpected status from node.js: %q", resp.Status)
	}
}

// CallContract overrides the same method in *ethclient.Client (and satisfies CallContract from
// go-ethereum's bind.ContractCaller interface).  Instead of sending the call through the
// underlying client, it sends it through 0x's library.
func (b *Backend) CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	var result string
	block := "latest"
	if blockNumber != nil {
		block = blockNumber.String()
	}

	convertedCall := map[string]interface{}{
		"data": hexutil.Encode(call.Data),
	}
	if call.From != (common.Address{}) {
		convertedCall["from"] = call.From
	}
	if call.To != nil {
		convertedCall["to"] = call.To
	}
	if call.Value != nil {
		convertedCall["value"] = call.Value
	}

	err := b.call(
		"call",
		map[string]interface{}{
			"call":  convertedCall,
			"block": block,
		},
		&result,
	)
	if err != nil {
		return nil, err
	}
	output, err := hex.DecodeString(strings.TrimPrefix(result, "0x"))
	if err != nil {
		return nil, err
	}
	return output, nil
}

// EstimateGas overrides the same method in *ethclient.Client (and satisfies EstimateGas from
// go-ethereum's bind.ContractTransactor interface). Instead of sending the call through the
// underlying client, it returns a hard-coded result.
//
// The hard-coded result is used so that transactions never fail in the gas estimation stage.
// Instead they will fail when the transaction is mined. This is assumed to be desirable behavior
// because it causes the transaction to actually run, rather than not, which causes the
// corresponding code to get traced for code coverage, rather than not.
func (b *Backend) EstimateGas(ctx context.Context, call ethereum.CallMsg) (gas uint64, err error) {
	return 8000000000, nil
}

// SendTransaction overrides the same method in *ethclient.Client (and satisfies SendTransaction
// from go-ethereum's bind.ContractTransactor interface). Instead of sending the call through the
// underlying client, it sends it through 0x's library.
func (b *Backend) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	buf := new(bytes.Buffer)
	err := tx.EncodeRLP(buf)
	if err != nil {
		return err
	}
	return b.call("sendTransaction", "0x"+hex.EncodeToString(buf.Bytes()), new(string) /* ignore output */)
}

// WriteCoverage writes a coverage report in Istanbul format to $PWD/coverage/coverage.json.
func (b *Backend) WriteCoverage() error {
	return b.call(
		"writeCoverage",
		true,      // ignored input
		new(bool), // ignore output
	)
}
