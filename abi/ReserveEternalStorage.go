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

// ReserveEternalStorageABI is the input ABI used to generate the binding from.
const ReserveEternalStorageABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"escapeHatch\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"key\",\"type\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"addBalance\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"from\",\"type\":\"address\"},{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"setAllowed\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"},{\"name\":\"\",\"type\":\"address\"}],\"name\":\"allowed\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newEscapeHatch\",\"type\":\"address\"}],\"name\":\"transferEscapeHatch\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"frozenTime\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"who\",\"type\":\"address\"},{\"name\":\"time\",\"type\":\"uint256\"}],\"name\":\"setFrozenTime\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"key\",\"type\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"subBalance\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"key\",\"type\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"setBalance\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"balance\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"escapeHatchAddress\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"oldOwner\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"oldEscapeHatch\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"newEscapeHatch\",\"type\":\"address\"}],\"name\":\"EscapeHatchTransferred\",\"type\":\"event\"}]"

// ReserveEternalStorageBin is the compiled bytecode used for deploying new contracts.
const ReserveEternalStorageBin = `608060405234801561001057600080fd5b50604051602080610bdc8339810180604052602081101561003057600080fd5b5051600080546001600160a01b03199081163317909155600180546001600160a01b0390931692909116919091179055610b6d8061006f6000396000f3fe608060405234801561001057600080fd5b50600436106100d45760003560e01c8063b062307411610081578063e30443bc1161005b578063e30443bc146102b5578063e3d670d7146102ee578063f2fde38b14610321576100d4565b8063b062307414610210578063b65dc41314610243578063cf8eeb7e1461027c576100d4565b80635c658165116100b25780635c658165146101885780638babf203146101d55780638da5cb5b14610208576100d4565b80631554611f146100d957806321e5383a1461010a57806333dd1b8a14610145575b600080fd5b6100e1610354565b6040805173ffffffffffffffffffffffffffffffffffffffff9092168252519081900360200190f35b6101436004803603604081101561012057600080fd5b5073ffffffffffffffffffffffffffffffffffffffff8135169060200135610370565b005b6101436004803603606081101561015b57600080fd5b5073ffffffffffffffffffffffffffffffffffffffff813581169160208101359091169060400135610459565b6101c36004803603604081101561019e57600080fd5b5073ffffffffffffffffffffffffffffffffffffffff81358116916020013516610518565b60408051918252519081900360200190f35b610143600480360360208110156101eb57600080fd5b503573ffffffffffffffffffffffffffffffffffffffff16610535565b6100e1610649565b6101c36004803603602081101561022657600080fd5b503573ffffffffffffffffffffffffffffffffffffffff16610665565b6101436004803603604081101561025957600080fd5b5073ffffffffffffffffffffffffffffffffffffffff8135169060200135610677565b6101436004803603604081101561029257600080fd5b5073ffffffffffffffffffffffffffffffffffffffff8135169060200135610726565b610143600480360360408110156102cb57600080fd5b5073ffffffffffffffffffffffffffffffffffffffff81351690602001356107e2565b6101c36004803603602081101561030457600080fd5b503573ffffffffffffffffffffffffffffffffffffffff16610891565b6101436004803603602081101561033757600080fd5b503573ffffffffffffffffffffffffffffffffffffffff166108a3565b60015473ffffffffffffffffffffffffffffffffffffffff1681565b60005473ffffffffffffffffffffffffffffffffffffffff1633146103f657604080517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152600960248201527f6f6e6c794f776e65720000000000000000000000000000000000000000000000604482015290519081900360640190fd5b73ffffffffffffffffffffffffffffffffffffffff821660009081526002602052604090205461042c908263ffffffff6109d816565b73ffffffffffffffffffffffffffffffffffffffff90921660009081526002602052604090209190915550565b60005473ffffffffffffffffffffffffffffffffffffffff1633146104df57604080517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152600960248201527f6f6e6c794f776e65720000000000000000000000000000000000000000000000604482015290519081900360640190fd5b73ffffffffffffffffffffffffffffffffffffffff92831660009081526003602090815260408083209490951682529290925291902055565b600360209081526000928352604080842090915290825290205481565b60015473ffffffffffffffffffffffffffffffffffffffff1633146105bb57604080517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152600e60248201527f6e6f7420617574686f72697a6564000000000000000000000000000000000000604482015290519081900360640190fd5b60015460405173ffffffffffffffffffffffffffffffffffffffff8084169216907f089af7288b55770a7c1dfd40b9d9e464c64031c45326c0916854814b6c16da2890600090a3600180547fffffffffffffffffffffffff00000000000000000000000000000000000000001673ffffffffffffffffffffffffffffffffffffffff92909216919091179055565b60005473ffffffffffffffffffffffffffffffffffffffff1681565b60046020526000908152604090205481565b60005473ffffffffffffffffffffffffffffffffffffffff1633146106fd57604080517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152600960248201527f6f6e6c794f776e65720000000000000000000000000000000000000000000000604482015290519081900360640190fd5b73ffffffffffffffffffffffffffffffffffffffff909116600090815260046020526040902055565b60005473ffffffffffffffffffffffffffffffffffffffff1633146107ac57604080517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152600960248201527f6f6e6c794f776e65720000000000000000000000000000000000000000000000604482015290519081900360640190fd5b73ffffffffffffffffffffffffffffffffffffffff821660009081526002602052604090205461042c908263ffffffff610a5316565b60005473ffffffffffffffffffffffffffffffffffffffff16331461086857604080517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152600960248201527f6f6e6c794f776e65720000000000000000000000000000000000000000000000604482015290519081900360640190fd5b73ffffffffffffffffffffffffffffffffffffffff909116600090815260026020526040902055565b60026020526000908152604090205481565b60005473ffffffffffffffffffffffffffffffffffffffff163314806108e0575060015473ffffffffffffffffffffffffffffffffffffffff1633145b61094b57604080517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152600e60248201527f6e6f7420617574686f72697a6564000000000000000000000000000000000000604482015290519081900360640190fd5b6000805460405173ffffffffffffffffffffffffffffffffffffffff808516939216917f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e091a3600080547fffffffffffffffffffffffff00000000000000000000000000000000000000001673ffffffffffffffffffffffffffffffffffffffff92909216919091179055565b600082820183811015610a4c57604080517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601b60248201527f536166654d6174683a206164646974696f6e206f766572666c6f770000000000604482015290519081900360640190fd5b9392505050565b6000610a4c83836040518060400160405280601e81526020017f536166654d6174683a207375627472616374696f6e206f766572666c6f77000081525060008184841115610b39576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825283818151815260200191508051906020019080838360005b83811015610afe578181015183820152602001610ae6565b50505050905090810190601f168015610b2b5780820380516001836020036101000a031916815260200191505b509250505060405180910390fd5b50505090039056fea165627a7a72305820b219396b330d0ef52dc2f277f2a517c53c78324478a1c84d3d5747b10635bd660029`

// DeployReserveEternalStorage deploys a new Ethereum contract, binding an instance of ReserveEternalStorage to it.
func DeployReserveEternalStorage(auth *bind.TransactOpts, backend bind.ContractBackend, escapeHatchAddress common.Address) (common.Address, *types.Transaction, *ReserveEternalStorage, error) {
	parsed, err := abi.JSON(strings.NewReader(ReserveEternalStorageABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(ReserveEternalStorageBin), backend, escapeHatchAddress)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &ReserveEternalStorage{ReserveEternalStorageCaller: ReserveEternalStorageCaller{contract: contract}, ReserveEternalStorageTransactor: ReserveEternalStorageTransactor{contract: contract}, ReserveEternalStorageFilterer: ReserveEternalStorageFilterer{contract: contract}}, nil
}

// ReserveEternalStorage is an auto generated Go binding around an Ethereum contract.
type ReserveEternalStorage struct {
	ReserveEternalStorageCaller     // Read-only binding to the contract
	ReserveEternalStorageTransactor // Write-only binding to the contract
	ReserveEternalStorageFilterer   // Log filterer for contract events
}

// ReserveEternalStorageCaller is an auto generated read-only Go binding around an Ethereum contract.
type ReserveEternalStorageCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ReserveEternalStorageTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ReserveEternalStorageTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ReserveEternalStorageFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ReserveEternalStorageFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ReserveEternalStorageSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ReserveEternalStorageSession struct {
	Contract     *ReserveEternalStorage // Generic contract binding to set the session for
	CallOpts     bind.CallOpts          // Call options to use throughout this session
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// ReserveEternalStorageCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ReserveEternalStorageCallerSession struct {
	Contract *ReserveEternalStorageCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                // Call options to use throughout this session
}

// ReserveEternalStorageTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ReserveEternalStorageTransactorSession struct {
	Contract     *ReserveEternalStorageTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                // Transaction auth options to use throughout this session
}

// ReserveEternalStorageRaw is an auto generated low-level Go binding around an Ethereum contract.
type ReserveEternalStorageRaw struct {
	Contract *ReserveEternalStorage // Generic contract binding to access the raw methods on
}

// ReserveEternalStorageCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ReserveEternalStorageCallerRaw struct {
	Contract *ReserveEternalStorageCaller // Generic read-only contract binding to access the raw methods on
}

// ReserveEternalStorageTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ReserveEternalStorageTransactorRaw struct {
	Contract *ReserveEternalStorageTransactor // Generic write-only contract binding to access the raw methods on
}

// NewReserveEternalStorage creates a new instance of ReserveEternalStorage, bound to a specific deployed contract.
func NewReserveEternalStorage(address common.Address, backend bind.ContractBackend) (*ReserveEternalStorage, error) {
	contract, err := bindReserveEternalStorage(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ReserveEternalStorage{ReserveEternalStorageCaller: ReserveEternalStorageCaller{contract: contract}, ReserveEternalStorageTransactor: ReserveEternalStorageTransactor{contract: contract}, ReserveEternalStorageFilterer: ReserveEternalStorageFilterer{contract: contract}}, nil
}

// NewReserveEternalStorageCaller creates a new read-only instance of ReserveEternalStorage, bound to a specific deployed contract.
func NewReserveEternalStorageCaller(address common.Address, caller bind.ContractCaller) (*ReserveEternalStorageCaller, error) {
	contract, err := bindReserveEternalStorage(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ReserveEternalStorageCaller{contract: contract}, nil
}

// NewReserveEternalStorageTransactor creates a new write-only instance of ReserveEternalStorage, bound to a specific deployed contract.
func NewReserveEternalStorageTransactor(address common.Address, transactor bind.ContractTransactor) (*ReserveEternalStorageTransactor, error) {
	contract, err := bindReserveEternalStorage(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ReserveEternalStorageTransactor{contract: contract}, nil
}

// NewReserveEternalStorageFilterer creates a new log filterer instance of ReserveEternalStorage, bound to a specific deployed contract.
func NewReserveEternalStorageFilterer(address common.Address, filterer bind.ContractFilterer) (*ReserveEternalStorageFilterer, error) {
	contract, err := bindReserveEternalStorage(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ReserveEternalStorageFilterer{contract: contract}, nil
}

// bindReserveEternalStorage binds a generic wrapper to an already deployed contract.
func bindReserveEternalStorage(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ReserveEternalStorageABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ReserveEternalStorage *ReserveEternalStorageRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _ReserveEternalStorage.Contract.ReserveEternalStorageCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ReserveEternalStorage *ReserveEternalStorageRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ReserveEternalStorage.Contract.ReserveEternalStorageTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ReserveEternalStorage *ReserveEternalStorageRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ReserveEternalStorage.Contract.ReserveEternalStorageTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ReserveEternalStorage *ReserveEternalStorageCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _ReserveEternalStorage.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ReserveEternalStorage *ReserveEternalStorageTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ReserveEternalStorage.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ReserveEternalStorage *ReserveEternalStorageTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ReserveEternalStorage.Contract.contract.Transact(opts, method, params...)
}

// Allowed is a free data retrieval call binding the contract method 0x5c658165.
//
// Solidity: function allowed(address , address ) constant returns(uint256)
func (_ReserveEternalStorage *ReserveEternalStorageCaller) Allowed(opts *bind.CallOpts, arg0 common.Address, arg1 common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _ReserveEternalStorage.contract.Call(opts, out, "allowed", arg0, arg1)
	return *ret0, err
}

// Allowed is a free data retrieval call binding the contract method 0x5c658165.
//
// Solidity: function allowed(address , address ) constant returns(uint256)
func (_ReserveEternalStorage *ReserveEternalStorageSession) Allowed(arg0 common.Address, arg1 common.Address) (*big.Int, error) {
	return _ReserveEternalStorage.Contract.Allowed(&_ReserveEternalStorage.CallOpts, arg0, arg1)
}

// Allowed is a free data retrieval call binding the contract method 0x5c658165.
//
// Solidity: function allowed(address , address ) constant returns(uint256)
func (_ReserveEternalStorage *ReserveEternalStorageCallerSession) Allowed(arg0 common.Address, arg1 common.Address) (*big.Int, error) {
	return _ReserveEternalStorage.Contract.Allowed(&_ReserveEternalStorage.CallOpts, arg0, arg1)
}

// Balance is a free data retrieval call binding the contract method 0xe3d670d7.
//
// Solidity: function balance(address ) constant returns(uint256)
func (_ReserveEternalStorage *ReserveEternalStorageCaller) Balance(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _ReserveEternalStorage.contract.Call(opts, out, "balance", arg0)
	return *ret0, err
}

// Balance is a free data retrieval call binding the contract method 0xe3d670d7.
//
// Solidity: function balance(address ) constant returns(uint256)
func (_ReserveEternalStorage *ReserveEternalStorageSession) Balance(arg0 common.Address) (*big.Int, error) {
	return _ReserveEternalStorage.Contract.Balance(&_ReserveEternalStorage.CallOpts, arg0)
}

// Balance is a free data retrieval call binding the contract method 0xe3d670d7.
//
// Solidity: function balance(address ) constant returns(uint256)
func (_ReserveEternalStorage *ReserveEternalStorageCallerSession) Balance(arg0 common.Address) (*big.Int, error) {
	return _ReserveEternalStorage.Contract.Balance(&_ReserveEternalStorage.CallOpts, arg0)
}

// EscapeHatch is a free data retrieval call binding the contract method 0x1554611f.
//
// Solidity: function escapeHatch() constant returns(address)
func (_ReserveEternalStorage *ReserveEternalStorageCaller) EscapeHatch(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _ReserveEternalStorage.contract.Call(opts, out, "escapeHatch")
	return *ret0, err
}

// EscapeHatch is a free data retrieval call binding the contract method 0x1554611f.
//
// Solidity: function escapeHatch() constant returns(address)
func (_ReserveEternalStorage *ReserveEternalStorageSession) EscapeHatch() (common.Address, error) {
	return _ReserveEternalStorage.Contract.EscapeHatch(&_ReserveEternalStorage.CallOpts)
}

// EscapeHatch is a free data retrieval call binding the contract method 0x1554611f.
//
// Solidity: function escapeHatch() constant returns(address)
func (_ReserveEternalStorage *ReserveEternalStorageCallerSession) EscapeHatch() (common.Address, error) {
	return _ReserveEternalStorage.Contract.EscapeHatch(&_ReserveEternalStorage.CallOpts)
}

// FrozenTime is a free data retrieval call binding the contract method 0xb0623074.
//
// Solidity: function frozenTime(address ) constant returns(uint256)
func (_ReserveEternalStorage *ReserveEternalStorageCaller) FrozenTime(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _ReserveEternalStorage.contract.Call(opts, out, "frozenTime", arg0)
	return *ret0, err
}

// FrozenTime is a free data retrieval call binding the contract method 0xb0623074.
//
// Solidity: function frozenTime(address ) constant returns(uint256)
func (_ReserveEternalStorage *ReserveEternalStorageSession) FrozenTime(arg0 common.Address) (*big.Int, error) {
	return _ReserveEternalStorage.Contract.FrozenTime(&_ReserveEternalStorage.CallOpts, arg0)
}

// FrozenTime is a free data retrieval call binding the contract method 0xb0623074.
//
// Solidity: function frozenTime(address ) constant returns(uint256)
func (_ReserveEternalStorage *ReserveEternalStorageCallerSession) FrozenTime(arg0 common.Address) (*big.Int, error) {
	return _ReserveEternalStorage.Contract.FrozenTime(&_ReserveEternalStorage.CallOpts, arg0)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_ReserveEternalStorage *ReserveEternalStorageCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _ReserveEternalStorage.contract.Call(opts, out, "owner")
	return *ret0, err
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_ReserveEternalStorage *ReserveEternalStorageSession) Owner() (common.Address, error) {
	return _ReserveEternalStorage.Contract.Owner(&_ReserveEternalStorage.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() constant returns(address)
func (_ReserveEternalStorage *ReserveEternalStorageCallerSession) Owner() (common.Address, error) {
	return _ReserveEternalStorage.Contract.Owner(&_ReserveEternalStorage.CallOpts)
}

// AddBalance is a paid mutator transaction binding the contract method 0x21e5383a.
//
// Solidity: function addBalance(address key, uint256 value) returns()
func (_ReserveEternalStorage *ReserveEternalStorageTransactor) AddBalance(opts *bind.TransactOpts, key common.Address, value *big.Int) (*types.Transaction, error) {
	return _ReserveEternalStorage.contract.Transact(opts, "addBalance", key, value)
}

// AddBalance is a paid mutator transaction binding the contract method 0x21e5383a.
//
// Solidity: function addBalance(address key, uint256 value) returns()
func (_ReserveEternalStorage *ReserveEternalStorageSession) AddBalance(key common.Address, value *big.Int) (*types.Transaction, error) {
	return _ReserveEternalStorage.Contract.AddBalance(&_ReserveEternalStorage.TransactOpts, key, value)
}

// AddBalance is a paid mutator transaction binding the contract method 0x21e5383a.
//
// Solidity: function addBalance(address key, uint256 value) returns()
func (_ReserveEternalStorage *ReserveEternalStorageTransactorSession) AddBalance(key common.Address, value *big.Int) (*types.Transaction, error) {
	return _ReserveEternalStorage.Contract.AddBalance(&_ReserveEternalStorage.TransactOpts, key, value)
}

// SetAllowed is a paid mutator transaction binding the contract method 0x33dd1b8a.
//
// Solidity: function setAllowed(address from, address to, uint256 value) returns()
func (_ReserveEternalStorage *ReserveEternalStorageTransactor) SetAllowed(opts *bind.TransactOpts, from common.Address, to common.Address, value *big.Int) (*types.Transaction, error) {
	return _ReserveEternalStorage.contract.Transact(opts, "setAllowed", from, to, value)
}

// SetAllowed is a paid mutator transaction binding the contract method 0x33dd1b8a.
//
// Solidity: function setAllowed(address from, address to, uint256 value) returns()
func (_ReserveEternalStorage *ReserveEternalStorageSession) SetAllowed(from common.Address, to common.Address, value *big.Int) (*types.Transaction, error) {
	return _ReserveEternalStorage.Contract.SetAllowed(&_ReserveEternalStorage.TransactOpts, from, to, value)
}

// SetAllowed is a paid mutator transaction binding the contract method 0x33dd1b8a.
//
// Solidity: function setAllowed(address from, address to, uint256 value) returns()
func (_ReserveEternalStorage *ReserveEternalStorageTransactorSession) SetAllowed(from common.Address, to common.Address, value *big.Int) (*types.Transaction, error) {
	return _ReserveEternalStorage.Contract.SetAllowed(&_ReserveEternalStorage.TransactOpts, from, to, value)
}

// SetBalance is a paid mutator transaction binding the contract method 0xe30443bc.
//
// Solidity: function setBalance(address key, uint256 value) returns()
func (_ReserveEternalStorage *ReserveEternalStorageTransactor) SetBalance(opts *bind.TransactOpts, key common.Address, value *big.Int) (*types.Transaction, error) {
	return _ReserveEternalStorage.contract.Transact(opts, "setBalance", key, value)
}

// SetBalance is a paid mutator transaction binding the contract method 0xe30443bc.
//
// Solidity: function setBalance(address key, uint256 value) returns()
func (_ReserveEternalStorage *ReserveEternalStorageSession) SetBalance(key common.Address, value *big.Int) (*types.Transaction, error) {
	return _ReserveEternalStorage.Contract.SetBalance(&_ReserveEternalStorage.TransactOpts, key, value)
}

// SetBalance is a paid mutator transaction binding the contract method 0xe30443bc.
//
// Solidity: function setBalance(address key, uint256 value) returns()
func (_ReserveEternalStorage *ReserveEternalStorageTransactorSession) SetBalance(key common.Address, value *big.Int) (*types.Transaction, error) {
	return _ReserveEternalStorage.Contract.SetBalance(&_ReserveEternalStorage.TransactOpts, key, value)
}

// SetFrozenTime is a paid mutator transaction binding the contract method 0xb65dc413.
//
// Solidity: function setFrozenTime(address who, uint256 time) returns()
func (_ReserveEternalStorage *ReserveEternalStorageTransactor) SetFrozenTime(opts *bind.TransactOpts, who common.Address, time *big.Int) (*types.Transaction, error) {
	return _ReserveEternalStorage.contract.Transact(opts, "setFrozenTime", who, time)
}

// SetFrozenTime is a paid mutator transaction binding the contract method 0xb65dc413.
//
// Solidity: function setFrozenTime(address who, uint256 time) returns()
func (_ReserveEternalStorage *ReserveEternalStorageSession) SetFrozenTime(who common.Address, time *big.Int) (*types.Transaction, error) {
	return _ReserveEternalStorage.Contract.SetFrozenTime(&_ReserveEternalStorage.TransactOpts, who, time)
}

// SetFrozenTime is a paid mutator transaction binding the contract method 0xb65dc413.
//
// Solidity: function setFrozenTime(address who, uint256 time) returns()
func (_ReserveEternalStorage *ReserveEternalStorageTransactorSession) SetFrozenTime(who common.Address, time *big.Int) (*types.Transaction, error) {
	return _ReserveEternalStorage.Contract.SetFrozenTime(&_ReserveEternalStorage.TransactOpts, who, time)
}

// SubBalance is a paid mutator transaction binding the contract method 0xcf8eeb7e.
//
// Solidity: function subBalance(address key, uint256 value) returns()
func (_ReserveEternalStorage *ReserveEternalStorageTransactor) SubBalance(opts *bind.TransactOpts, key common.Address, value *big.Int) (*types.Transaction, error) {
	return _ReserveEternalStorage.contract.Transact(opts, "subBalance", key, value)
}

// SubBalance is a paid mutator transaction binding the contract method 0xcf8eeb7e.
//
// Solidity: function subBalance(address key, uint256 value) returns()
func (_ReserveEternalStorage *ReserveEternalStorageSession) SubBalance(key common.Address, value *big.Int) (*types.Transaction, error) {
	return _ReserveEternalStorage.Contract.SubBalance(&_ReserveEternalStorage.TransactOpts, key, value)
}

// SubBalance is a paid mutator transaction binding the contract method 0xcf8eeb7e.
//
// Solidity: function subBalance(address key, uint256 value) returns()
func (_ReserveEternalStorage *ReserveEternalStorageTransactorSession) SubBalance(key common.Address, value *big.Int) (*types.Transaction, error) {
	return _ReserveEternalStorage.Contract.SubBalance(&_ReserveEternalStorage.TransactOpts, key, value)
}

// TransferEscapeHatch is a paid mutator transaction binding the contract method 0x8babf203.
//
// Solidity: function transferEscapeHatch(address newEscapeHatch) returns()
func (_ReserveEternalStorage *ReserveEternalStorageTransactor) TransferEscapeHatch(opts *bind.TransactOpts, newEscapeHatch common.Address) (*types.Transaction, error) {
	return _ReserveEternalStorage.contract.Transact(opts, "transferEscapeHatch", newEscapeHatch)
}

// TransferEscapeHatch is a paid mutator transaction binding the contract method 0x8babf203.
//
// Solidity: function transferEscapeHatch(address newEscapeHatch) returns()
func (_ReserveEternalStorage *ReserveEternalStorageSession) TransferEscapeHatch(newEscapeHatch common.Address) (*types.Transaction, error) {
	return _ReserveEternalStorage.Contract.TransferEscapeHatch(&_ReserveEternalStorage.TransactOpts, newEscapeHatch)
}

// TransferEscapeHatch is a paid mutator transaction binding the contract method 0x8babf203.
//
// Solidity: function transferEscapeHatch(address newEscapeHatch) returns()
func (_ReserveEternalStorage *ReserveEternalStorageTransactorSession) TransferEscapeHatch(newEscapeHatch common.Address) (*types.Transaction, error) {
	return _ReserveEternalStorage.Contract.TransferEscapeHatch(&_ReserveEternalStorage.TransactOpts, newEscapeHatch)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ReserveEternalStorage *ReserveEternalStorageTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _ReserveEternalStorage.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ReserveEternalStorage *ReserveEternalStorageSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _ReserveEternalStorage.Contract.TransferOwnership(&_ReserveEternalStorage.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ReserveEternalStorage *ReserveEternalStorageTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _ReserveEternalStorage.Contract.TransferOwnership(&_ReserveEternalStorage.TransactOpts, newOwner)
}

// ReserveEternalStorageEscapeHatchTransferredIterator is returned from FilterEscapeHatchTransferred and is used to iterate over the raw logs and unpacked data for EscapeHatchTransferred events raised by the ReserveEternalStorage contract.
type ReserveEternalStorageEscapeHatchTransferredIterator struct {
	Event *ReserveEternalStorageEscapeHatchTransferred // Event containing the contract specifics and raw log

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
func (it *ReserveEternalStorageEscapeHatchTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ReserveEternalStorageEscapeHatchTransferred)
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
		it.Event = new(ReserveEternalStorageEscapeHatchTransferred)
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
func (it *ReserveEternalStorageEscapeHatchTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ReserveEternalStorageEscapeHatchTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ReserveEternalStorageEscapeHatchTransferred represents a EscapeHatchTransferred event raised by the ReserveEternalStorage contract.
type ReserveEternalStorageEscapeHatchTransferred struct {
	OldEscapeHatch common.Address
	NewEscapeHatch common.Address
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterEscapeHatchTransferred is a free log retrieval operation binding the contract event 0x089af7288b55770a7c1dfd40b9d9e464c64031c45326c0916854814b6c16da28.
//
// Solidity: event EscapeHatchTransferred(address indexed oldEscapeHatch, address indexed newEscapeHatch)
func (_ReserveEternalStorage *ReserveEternalStorageFilterer) FilterEscapeHatchTransferred(opts *bind.FilterOpts, oldEscapeHatch []common.Address, newEscapeHatch []common.Address) (*ReserveEternalStorageEscapeHatchTransferredIterator, error) {

	var oldEscapeHatchRule []interface{}
	for _, oldEscapeHatchItem := range oldEscapeHatch {
		oldEscapeHatchRule = append(oldEscapeHatchRule, oldEscapeHatchItem)
	}
	var newEscapeHatchRule []interface{}
	for _, newEscapeHatchItem := range newEscapeHatch {
		newEscapeHatchRule = append(newEscapeHatchRule, newEscapeHatchItem)
	}

	logs, sub, err := _ReserveEternalStorage.contract.FilterLogs(opts, "EscapeHatchTransferred", oldEscapeHatchRule, newEscapeHatchRule)
	if err != nil {
		return nil, err
	}
	return &ReserveEternalStorageEscapeHatchTransferredIterator{contract: _ReserveEternalStorage.contract, event: "EscapeHatchTransferred", logs: logs, sub: sub}, nil
}

// WatchEscapeHatchTransferred is a free log subscription operation binding the contract event 0x089af7288b55770a7c1dfd40b9d9e464c64031c45326c0916854814b6c16da28.
//
// Solidity: event EscapeHatchTransferred(address indexed oldEscapeHatch, address indexed newEscapeHatch)
func (_ReserveEternalStorage *ReserveEternalStorageFilterer) WatchEscapeHatchTransferred(opts *bind.WatchOpts, sink chan<- *ReserveEternalStorageEscapeHatchTransferred, oldEscapeHatch []common.Address, newEscapeHatch []common.Address) (event.Subscription, error) {

	var oldEscapeHatchRule []interface{}
	for _, oldEscapeHatchItem := range oldEscapeHatch {
		oldEscapeHatchRule = append(oldEscapeHatchRule, oldEscapeHatchItem)
	}
	var newEscapeHatchRule []interface{}
	for _, newEscapeHatchItem := range newEscapeHatch {
		newEscapeHatchRule = append(newEscapeHatchRule, newEscapeHatchItem)
	}

	logs, sub, err := _ReserveEternalStorage.contract.WatchLogs(opts, "EscapeHatchTransferred", oldEscapeHatchRule, newEscapeHatchRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ReserveEternalStorageEscapeHatchTransferred)
				if err := _ReserveEternalStorage.contract.UnpackLog(event, "EscapeHatchTransferred", log); err != nil {
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

// ReserveEternalStorageOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the ReserveEternalStorage contract.
type ReserveEternalStorageOwnershipTransferredIterator struct {
	Event *ReserveEternalStorageOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *ReserveEternalStorageOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ReserveEternalStorageOwnershipTransferred)
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
		it.Event = new(ReserveEternalStorageOwnershipTransferred)
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
func (it *ReserveEternalStorageOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ReserveEternalStorageOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ReserveEternalStorageOwnershipTransferred represents a OwnershipTransferred event raised by the ReserveEternalStorage contract.
type ReserveEternalStorageOwnershipTransferred struct {
	OldOwner common.Address
	NewOwner common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed oldOwner, address indexed newOwner)
func (_ReserveEternalStorage *ReserveEternalStorageFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, oldOwner []common.Address, newOwner []common.Address) (*ReserveEternalStorageOwnershipTransferredIterator, error) {

	var oldOwnerRule []interface{}
	for _, oldOwnerItem := range oldOwner {
		oldOwnerRule = append(oldOwnerRule, oldOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _ReserveEternalStorage.contract.FilterLogs(opts, "OwnershipTransferred", oldOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &ReserveEternalStorageOwnershipTransferredIterator{contract: _ReserveEternalStorage.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed oldOwner, address indexed newOwner)
func (_ReserveEternalStorage *ReserveEternalStorageFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *ReserveEternalStorageOwnershipTransferred, oldOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var oldOwnerRule []interface{}
	for _, oldOwnerItem := range oldOwner {
		oldOwnerRule = append(oldOwnerRule, oldOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _ReserveEternalStorage.contract.WatchLogs(opts, "OwnershipTransferred", oldOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ReserveEternalStorageOwnershipTransferred)
				if err := _ReserveEternalStorage.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
