// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package abi

import (
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = abi.U256
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// VaultABI is the input ABI used to generate the binding from.
const VaultABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"nominateNewOwner\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"manager\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"acceptOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"tokens\",\"type\":\"address[]\"},{\"name\":\"quantities\",\"type\":\"uint256[]\"},{\"name\":\"to\",\"type\":\"address\"}],\"name\":\"batchWithdrawTo\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newManager\",\"type\":\"address\"}],\"name\":\"changeManager\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"_nominatedOwner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"_owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"previousManager\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"newManager\",\"type\":\"address\"}],\"name\":\"ManagerTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"tokens\",\"type\":\"address[]\"},{\"indexed\":true,\"name\":\"quantities\",\"type\":\"uint256[]\"}],\"name\":\"BatchWithdrawal\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"NewOwnerNominated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"}]"

// VaultBin is the compiled bytecode used for deploying new contracts.
const VaultBin = `608060405234801561001057600080fd5b50600061002161008260201b60201c565b600080546001600160a01b0319166001600160a01b0383169081178255604051929350917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0908290a350600280546001600160a01b03191633179055610086565b3390565b610b56806100956000396000f3fe608060405234801561001057600080fd5b506004361061007d5760003560e01c80639ab1de9d1161005b5780639ab1de9d146100f0578063a3fbbaae146101c8578063aaf380f1146101fb578063b2bdfa7b146102035761007d565b80631627540c14610082578063481c6a75146100b757806379ba5097146100e8575b600080fd5b6100b56004803603602081101561009857600080fd5b503573ffffffffffffffffffffffffffffffffffffffff1661020b565b005b6100bf6103a7565b6040805173ffffffffffffffffffffffffffffffffffffffff9092168252519081900360200190f35b6100b56103c3565b6100b56004803603606081101561010657600080fd5b81019060208101813564010000000081111561012157600080fd5b82018360208201111561013357600080fd5b8035906020019184602083028401116401000000008311171561015557600080fd5b91939092909160208101903564010000000081111561017357600080fd5b82018360208201111561018557600080fd5b803590602001918460208302840111640100000000831117156101a757600080fd5b91935091503573ffffffffffffffffffffffffffffffffffffffff166104d1565b6100b5600480360360208110156101de57600080fd5b503573ffffffffffffffffffffffffffffffffffffffff16610654565b6100bf61078c565b6100bf6107a8565b60005473ffffffffffffffffffffffffffffffffffffffff1661022c6107c4565b73ffffffffffffffffffffffffffffffffffffffff16146102ae57604080517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b73ffffffffffffffffffffffffffffffffffffffff811661031a576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401808060200182810382526026815260200180610adb6026913960400191505060405180910390fd5b6000805460405173ffffffffffffffffffffffffffffffffffffffff808516939216917fb59bab42c554cfd49f4f001c983b6ed93ede25748b10114b7d1cb1b3c97df7af91a3600180547fffffffffffffffffffffffff00000000000000000000000000000000000000001673ffffffffffffffffffffffffffffffffffffffff92909216919091179055565b60025473ffffffffffffffffffffffffffffffffffffffff1681565b6103cb6107c4565b60015473ffffffffffffffffffffffffffffffffffffffff90811691161461043e576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401808060200182810382526026815260200180610adb6026913960400191505060405180910390fd5b6001546000805460405173ffffffffffffffffffffffffffffffffffffffff93841693909116917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e091a3600154600080547fffffffffffffffffffffffff00000000000000000000000000000000000000001673ffffffffffffffffffffffffffffffffffffffff909216919091179055565b60025473ffffffffffffffffffffffffffffffffffffffff16331461055757604080517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152600f60248201527f6d757374206265206d616e616765720000000000000000000000000000000000604482015290519081900360640190fd5b60005b848110156105ea57600084848381811061057057fe5b9050602002013511156105e2576105e28285858481811061058d57fe5b905060200201358888858181106105a057fe5b9050602002013573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff166107c89092919063ffffffff16565b60010161055a565b508282604051808383602002808284376040519201829003822094508993508892508190508360208402808284376040519201829003822094507f68b4558431231278e57e1a969c13c61605ca644c57a73d231bfcc7e2af80d2e093506000925050a35050505050565b60005473ffffffffffffffffffffffffffffffffffffffff166106756107c4565b73ffffffffffffffffffffffffffffffffffffffff16146106f757604080517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b73ffffffffffffffffffffffffffffffffffffffff811661071757600080fd5b600280547fffffffffffffffffffffffff00000000000000000000000000000000000000001673ffffffffffffffffffffffffffffffffffffffff83811691821792839055604051919216907f9cb45c728de594dab506a1f1a8554e24c8eeaf983618d5ec5dd7bc6f3c49feee90600090a350565b60015473ffffffffffffffffffffffffffffffffffffffff1681565b60005473ffffffffffffffffffffffffffffffffffffffff1681565b3390565b6040805173ffffffffffffffffffffffffffffffffffffffff8416602482015260448082018490528251808303909101815260649091019091526020810180517bffffffffffffffffffffffffffffffffffffffffffffffffffffffff167fa9059cbb0000000000000000000000000000000000000000000000000000000017905261085590849061085a565b505050565b6108798273ffffffffffffffffffffffffffffffffffffffff16610a9e565b6108e457604080517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601f60248201527f5361666545524332303a2063616c6c20746f206e6f6e2d636f6e747261637400604482015290519081900360640190fd5b600060608373ffffffffffffffffffffffffffffffffffffffff16836040518082805190602001908083835b6020831061094d57805182527fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe09092019160209182019101610910565b6001836020036101000a0380198251168184511680821785525050505050509050019150506000604051808303816000865af19150503d80600081146109af576040519150601f19603f3d011682016040523d82523d6000602084013e6109b4565b606091505b509150915081610a2557604080517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820181905260248201527f5361666545524332303a206c6f772d6c6576656c2063616c6c206661696c6564604482015290519081900360640190fd5b805115610a9857808060200190516020811015610a4157600080fd5b5051610a98576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040180806020018281038252602a815260200180610b01602a913960400191505060405180910390fd5b50505050565b6000813f7fc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a4708115801590610ad25750808214155b94935050505056fe4f776e61626c653a206e6577206f776e657220697320746865207a65726f20616464726573735361666545524332303a204552433230206f7065726174696f6e20646964206e6f742073756363656564a165627a7a72305820bbda696b14608ceef43860347b1d7c96382fc70fc4c7442b494d20dd5b54603e0029`

// DeployVault deploys a new Ethereum contract, binding an instance of Vault to it.
func DeployVault(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Vault, error) {
	parsed, err := abi.JSON(strings.NewReader(VaultABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(VaultBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Vault{VaultCaller: VaultCaller{contract: contract}, VaultTransactor: VaultTransactor{contract: contract}, VaultFilterer: VaultFilterer{contract: contract}}, nil
}

// Vault is an auto generated Go binding around an Ethereum contract.
type Vault struct {
	VaultCaller     // Read-only binding to the contract
	VaultTransactor // Write-only binding to the contract
	VaultFilterer   // Log filterer for contract events
}

// VaultCaller is an auto generated read-only Go binding around an Ethereum contract.
type VaultCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// VaultTransactor is an auto generated write-only Go binding around an Ethereum contract.
type VaultTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// VaultFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type VaultFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// VaultSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type VaultSession struct {
	Contract     *Vault            // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// VaultCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type VaultCallerSession struct {
	Contract *VaultCaller  // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// VaultTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type VaultTransactorSession struct {
	Contract     *VaultTransactor  // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// VaultRaw is an auto generated low-level Go binding around an Ethereum contract.
type VaultRaw struct {
	Contract *Vault // Generic contract binding to access the raw methods on
}

// VaultCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type VaultCallerRaw struct {
	Contract *VaultCaller // Generic read-only contract binding to access the raw methods on
}

// VaultTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type VaultTransactorRaw struct {
	Contract *VaultTransactor // Generic write-only contract binding to access the raw methods on
}

// NewVault creates a new instance of Vault, bound to a specific deployed contract.
func NewVault(address common.Address, backend bind.ContractBackend) (*Vault, error) {
	contract, err := bindVault(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Vault{VaultCaller: VaultCaller{contract: contract}, VaultTransactor: VaultTransactor{contract: contract}, VaultFilterer: VaultFilterer{contract: contract}}, nil
}

// NewVaultCaller creates a new read-only instance of Vault, bound to a specific deployed contract.
func NewVaultCaller(address common.Address, caller bind.ContractCaller) (*VaultCaller, error) {
	contract, err := bindVault(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &VaultCaller{contract: contract}, nil
}

// NewVaultTransactor creates a new write-only instance of Vault, bound to a specific deployed contract.
func NewVaultTransactor(address common.Address, transactor bind.ContractTransactor) (*VaultTransactor, error) {
	contract, err := bindVault(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &VaultTransactor{contract: contract}, nil
}

// NewVaultFilterer creates a new log filterer instance of Vault, bound to a specific deployed contract.
func NewVaultFilterer(address common.Address, filterer bind.ContractFilterer) (*VaultFilterer, error) {
	contract, err := bindVault(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &VaultFilterer{contract: contract}, nil
}

// bindVault binds a generic wrapper to an already deployed contract.
func bindVault(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(VaultABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Vault *VaultRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Vault.Contract.VaultCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Vault *VaultRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Vault.Contract.VaultTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Vault *VaultRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Vault.Contract.VaultTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Vault *VaultCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Vault.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Vault *VaultTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Vault.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Vault *VaultTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Vault.Contract.contract.Transact(opts, method, params...)
}

// NominatedOwner is a free data retrieval call binding the contract method 0xaaf380f1.
//
// Solidity: function _nominatedOwner() constant returns(address)
func (_Vault *VaultCaller) NominatedOwner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Vault.contract.Call(opts, out, "_nominatedOwner")
	return *ret0, err
}

// NominatedOwner is a free data retrieval call binding the contract method 0xaaf380f1.
//
// Solidity: function _nominatedOwner() constant returns(address)
func (_Vault *VaultSession) NominatedOwner() (common.Address, error) {
	return _Vault.Contract.NominatedOwner(&_Vault.CallOpts)
}

// NominatedOwner is a free data retrieval call binding the contract method 0xaaf380f1.
//
// Solidity: function _nominatedOwner() constant returns(address)
func (_Vault *VaultCallerSession) NominatedOwner() (common.Address, error) {
	return _Vault.Contract.NominatedOwner(&_Vault.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0xb2bdfa7b.
//
// Solidity: function _owner() constant returns(address)
func (_Vault *VaultCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Vault.contract.Call(opts, out, "_owner")
	return *ret0, err
}

// Owner is a free data retrieval call binding the contract method 0xb2bdfa7b.
//
// Solidity: function _owner() constant returns(address)
func (_Vault *VaultSession) Owner() (common.Address, error) {
	return _Vault.Contract.Owner(&_Vault.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0xb2bdfa7b.
//
// Solidity: function _owner() constant returns(address)
func (_Vault *VaultCallerSession) Owner() (common.Address, error) {
	return _Vault.Contract.Owner(&_Vault.CallOpts)
}

// Manager is a free data retrieval call binding the contract method 0x481c6a75.
//
// Solidity: function manager() constant returns(address)
func (_Vault *VaultCaller) Manager(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Vault.contract.Call(opts, out, "manager")
	return *ret0, err
}

// Manager is a free data retrieval call binding the contract method 0x481c6a75.
//
// Solidity: function manager() constant returns(address)
func (_Vault *VaultSession) Manager() (common.Address, error) {
	return _Vault.Contract.Manager(&_Vault.CallOpts)
}

// Manager is a free data retrieval call binding the contract method 0x481c6a75.
//
// Solidity: function manager() constant returns(address)
func (_Vault *VaultCallerSession) Manager() (common.Address, error) {
	return _Vault.Contract.Manager(&_Vault.CallOpts)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_Vault *VaultTransactor) AcceptOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Vault.contract.Transact(opts, "acceptOwnership")
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_Vault *VaultSession) AcceptOwnership() (*types.Transaction, error) {
	return _Vault.Contract.AcceptOwnership(&_Vault.TransactOpts)
}

// AcceptOwnership is a paid mutator transaction binding the contract method 0x79ba5097.
//
// Solidity: function acceptOwnership() returns()
func (_Vault *VaultTransactorSession) AcceptOwnership() (*types.Transaction, error) {
	return _Vault.Contract.AcceptOwnership(&_Vault.TransactOpts)
}

// BatchWithdrawTo is a paid mutator transaction binding the contract method 0x9ab1de9d.
//
// Solidity: function batchWithdrawTo(address[] tokens, uint256[] quantities, address to) returns()
func (_Vault *VaultTransactor) BatchWithdrawTo(opts *bind.TransactOpts, tokens []common.Address, quantities []*big.Int, to common.Address) (*types.Transaction, error) {
	return _Vault.contract.Transact(opts, "batchWithdrawTo", tokens, quantities, to)
}

// BatchWithdrawTo is a paid mutator transaction binding the contract method 0x9ab1de9d.
//
// Solidity: function batchWithdrawTo(address[] tokens, uint256[] quantities, address to) returns()
func (_Vault *VaultSession) BatchWithdrawTo(tokens []common.Address, quantities []*big.Int, to common.Address) (*types.Transaction, error) {
	return _Vault.Contract.BatchWithdrawTo(&_Vault.TransactOpts, tokens, quantities, to)
}

// BatchWithdrawTo is a paid mutator transaction binding the contract method 0x9ab1de9d.
//
// Solidity: function batchWithdrawTo(address[] tokens, uint256[] quantities, address to) returns()
func (_Vault *VaultTransactorSession) BatchWithdrawTo(tokens []common.Address, quantities []*big.Int, to common.Address) (*types.Transaction, error) {
	return _Vault.Contract.BatchWithdrawTo(&_Vault.TransactOpts, tokens, quantities, to)
}

// ChangeManager is a paid mutator transaction binding the contract method 0xa3fbbaae.
//
// Solidity: function changeManager(address newManager) returns()
func (_Vault *VaultTransactor) ChangeManager(opts *bind.TransactOpts, newManager common.Address) (*types.Transaction, error) {
	return _Vault.contract.Transact(opts, "changeManager", newManager)
}

// ChangeManager is a paid mutator transaction binding the contract method 0xa3fbbaae.
//
// Solidity: function changeManager(address newManager) returns()
func (_Vault *VaultSession) ChangeManager(newManager common.Address) (*types.Transaction, error) {
	return _Vault.Contract.ChangeManager(&_Vault.TransactOpts, newManager)
}

// ChangeManager is a paid mutator transaction binding the contract method 0xa3fbbaae.
//
// Solidity: function changeManager(address newManager) returns()
func (_Vault *VaultTransactorSession) ChangeManager(newManager common.Address) (*types.Transaction, error) {
	return _Vault.Contract.ChangeManager(&_Vault.TransactOpts, newManager)
}

// NominateNewOwner is a paid mutator transaction binding the contract method 0x1627540c.
//
// Solidity: function nominateNewOwner(address newOwner) returns()
func (_Vault *VaultTransactor) NominateNewOwner(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Vault.contract.Transact(opts, "nominateNewOwner", newOwner)
}

// NominateNewOwner is a paid mutator transaction binding the contract method 0x1627540c.
//
// Solidity: function nominateNewOwner(address newOwner) returns()
func (_Vault *VaultSession) NominateNewOwner(newOwner common.Address) (*types.Transaction, error) {
	return _Vault.Contract.NominateNewOwner(&_Vault.TransactOpts, newOwner)
}

// NominateNewOwner is a paid mutator transaction binding the contract method 0x1627540c.
//
// Solidity: function nominateNewOwner(address newOwner) returns()
func (_Vault *VaultTransactorSession) NominateNewOwner(newOwner common.Address) (*types.Transaction, error) {
	return _Vault.Contract.NominateNewOwner(&_Vault.TransactOpts, newOwner)
}

// VaultBatchWithdrawalIterator is returned from FilterBatchWithdrawal and is used to iterate over the raw logs and unpacked data for BatchWithdrawal events raised by the Vault contract.
type VaultBatchWithdrawalIterator struct {
	Event *VaultBatchWithdrawal // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *VaultBatchWithdrawalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(VaultBatchWithdrawal)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(VaultBatchWithdrawal)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *VaultBatchWithdrawalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *VaultBatchWithdrawalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// VaultBatchWithdrawal represents a BatchWithdrawal event raised by the Vault contract.
type VaultBatchWithdrawal struct {
	Tokens     []common.Address
	Quantities []*big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterBatchWithdrawal is a free log retrieval operation binding the contract event 0x68b4558431231278e57e1a969c13c61605ca644c57a73d231bfcc7e2af80d2e0.
//
// Solidity: event BatchWithdrawal(address[] indexed tokens, uint256[] indexed quantities)
func (_Vault *VaultFilterer) FilterBatchWithdrawal(opts *bind.FilterOpts, tokens [][]common.Address, quantities [][]*big.Int) (*VaultBatchWithdrawalIterator, error) {

	var tokensRule []interface{}
	for _, tokensItem := range tokens {
		tokensRule = append(tokensRule, tokensItem)
	}
	var quantitiesRule []interface{}
	for _, quantitiesItem := range quantities {
		quantitiesRule = append(quantitiesRule, quantitiesItem)
	}

	logs, sub, err := _Vault.contract.FilterLogs(opts, "BatchWithdrawal", tokensRule, quantitiesRule)
	if err != nil {
		return nil, err
	}
	return &VaultBatchWithdrawalIterator{contract: _Vault.contract, event: "BatchWithdrawal", logs: logs, sub: sub}, nil
}

// WatchBatchWithdrawal is a free log subscription operation binding the contract event 0x68b4558431231278e57e1a969c13c61605ca644c57a73d231bfcc7e2af80d2e0.
//
// Solidity: event BatchWithdrawal(address[] indexed tokens, uint256[] indexed quantities)
func (_Vault *VaultFilterer) WatchBatchWithdrawal(opts *bind.WatchOpts, sink chan<- *VaultBatchWithdrawal, tokens [][]common.Address, quantities [][]*big.Int) (event.Subscription, error) {

	var tokensRule []interface{}
	for _, tokensItem := range tokens {
		tokensRule = append(tokensRule, tokensItem)
	}
	var quantitiesRule []interface{}
	for _, quantitiesItem := range quantities {
		quantitiesRule = append(quantitiesRule, quantitiesItem)
	}

	logs, sub, err := _Vault.contract.WatchLogs(opts, "BatchWithdrawal", tokensRule, quantitiesRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(VaultBatchWithdrawal)
				if err := _Vault.contract.UnpackLog(event, "BatchWithdrawal", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// VaultManagerTransferredIterator is returned from FilterManagerTransferred and is used to iterate over the raw logs and unpacked data for ManagerTransferred events raised by the Vault contract.
type VaultManagerTransferredIterator struct {
	Event *VaultManagerTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *VaultManagerTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(VaultManagerTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(VaultManagerTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *VaultManagerTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *VaultManagerTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// VaultManagerTransferred represents a ManagerTransferred event raised by the Vault contract.
type VaultManagerTransferred struct {
	PreviousManager common.Address
	NewManager      common.Address
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterManagerTransferred is a free log retrieval operation binding the contract event 0x9cb45c728de594dab506a1f1a8554e24c8eeaf983618d5ec5dd7bc6f3c49feee.
//
// Solidity: event ManagerTransferred(address indexed previousManager, address indexed newManager)
func (_Vault *VaultFilterer) FilterManagerTransferred(opts *bind.FilterOpts, previousManager []common.Address, newManager []common.Address) (*VaultManagerTransferredIterator, error) {

	var previousManagerRule []interface{}
	for _, previousManagerItem := range previousManager {
		previousManagerRule = append(previousManagerRule, previousManagerItem)
	}
	var newManagerRule []interface{}
	for _, newManagerItem := range newManager {
		newManagerRule = append(newManagerRule, newManagerItem)
	}

	logs, sub, err := _Vault.contract.FilterLogs(opts, "ManagerTransferred", previousManagerRule, newManagerRule)
	if err != nil {
		return nil, err
	}
	return &VaultManagerTransferredIterator{contract: _Vault.contract, event: "ManagerTransferred", logs: logs, sub: sub}, nil
}

// WatchManagerTransferred is a free log subscription operation binding the contract event 0x9cb45c728de594dab506a1f1a8554e24c8eeaf983618d5ec5dd7bc6f3c49feee.
//
// Solidity: event ManagerTransferred(address indexed previousManager, address indexed newManager)
func (_Vault *VaultFilterer) WatchManagerTransferred(opts *bind.WatchOpts, sink chan<- *VaultManagerTransferred, previousManager []common.Address, newManager []common.Address) (event.Subscription, error) {

	var previousManagerRule []interface{}
	for _, previousManagerItem := range previousManager {
		previousManagerRule = append(previousManagerRule, previousManagerItem)
	}
	var newManagerRule []interface{}
	for _, newManagerItem := range newManager {
		newManagerRule = append(newManagerRule, newManagerItem)
	}

	logs, sub, err := _Vault.contract.WatchLogs(opts, "ManagerTransferred", previousManagerRule, newManagerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(VaultManagerTransferred)
				if err := _Vault.contract.UnpackLog(event, "ManagerTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// VaultNewOwnerNominatedIterator is returned from FilterNewOwnerNominated and is used to iterate over the raw logs and unpacked data for NewOwnerNominated events raised by the Vault contract.
type VaultNewOwnerNominatedIterator struct {
	Event *VaultNewOwnerNominated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *VaultNewOwnerNominatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(VaultNewOwnerNominated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(VaultNewOwnerNominated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *VaultNewOwnerNominatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *VaultNewOwnerNominatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// VaultNewOwnerNominated represents a NewOwnerNominated event raised by the Vault contract.
type VaultNewOwnerNominated struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterNewOwnerNominated is a free log retrieval operation binding the contract event 0xb59bab42c554cfd49f4f001c983b6ed93ede25748b10114b7d1cb1b3c97df7af.
//
// Solidity: event NewOwnerNominated(address indexed previousOwner, address indexed newOwner)
func (_Vault *VaultFilterer) FilterNewOwnerNominated(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*VaultNewOwnerNominatedIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Vault.contract.FilterLogs(opts, "NewOwnerNominated", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &VaultNewOwnerNominatedIterator{contract: _Vault.contract, event: "NewOwnerNominated", logs: logs, sub: sub}, nil
}

// WatchNewOwnerNominated is a free log subscription operation binding the contract event 0xb59bab42c554cfd49f4f001c983b6ed93ede25748b10114b7d1cb1b3c97df7af.
//
// Solidity: event NewOwnerNominated(address indexed previousOwner, address indexed newOwner)
func (_Vault *VaultFilterer) WatchNewOwnerNominated(opts *bind.WatchOpts, sink chan<- *VaultNewOwnerNominated, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Vault.contract.WatchLogs(opts, "NewOwnerNominated", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(VaultNewOwnerNominated)
				if err := _Vault.contract.UnpackLog(event, "NewOwnerNominated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// VaultOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the Vault contract.
type VaultOwnershipTransferredIterator struct {
	Event *VaultOwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *VaultOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(VaultOwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(VaultOwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *VaultOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *VaultOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// VaultOwnershipTransferred represents a OwnershipTransferred event raised by the Vault contract.
type VaultOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Vault *VaultFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*VaultOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Vault.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &VaultOwnershipTransferredIterator{contract: _Vault.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Vault *VaultFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *VaultOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Vault.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(VaultOwnershipTransferred)
				if err := _Vault.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}
