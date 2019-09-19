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

// BasketABI is the input ABI used to generate the binding from.
const BasketABI = "[{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"backing\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_frontTokenSupply\",\"type\":\"uint256\"},{\"name\":\"_other\",\"type\":\"address\"}],\"name\":\"newQuantitiesRequired\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"tokens\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"\",\"type\":\"address\"}],\"name\":\"backingMap\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"size\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"frontTokenDecimals\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getTokens\",\"outputs\":[{\"name\":\"\",\"type\":\"address[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"_frontTokenSupply\",\"type\":\"uint256\"}],\"name\":\"quantitiesRequired\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_tokens\",\"type\":\"address[]\"},{\"name\":\"_backing\",\"type\":\"uint256[]\"},{\"name\":\"_frontTokenDecimals\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"}]"

// BasketBin is the compiled bytecode used for deploying new contracts.
const BasketBin = `60806040523480156200001157600080fd5b5060405162000c8338038062000c83833981018060405260608110156200003757600080fd5b8101908080516401000000008111156200005057600080fd5b820160208101848111156200006457600080fd5b81518560208202830111640100000000821117156200008257600080fd5b505092919060200180516401000000008111156200009f57600080fd5b82016020810184811115620000b357600080fd5b8151856020820283011164010000000082111715620000d157600080fd5b505060209091015181518551929450909250146200015057604080517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152600e60248201527f696e76616c6964206261736b6574000000000000000000000000000000000000604482015290519081900360640190fd5b6000835111620001c157604080517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601060248201527f6261736b657420746f6f20736d616c6c00000000000000000000000000000000604482015290519081900360640190fd5b6103e8835111156200023457604080517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152600e60248201527f6261736b657420746f6f20626967000000000000000000000000000000000000604482015290519081900360640190fd5b8251620002499060019060208601906200026c565b5081516200025f906002906020850190620002d6565b5060005550620003699050565b828054828255906000526020600020908101928215620002c4579160200282015b82811115620002c457825182546001600160a01b0319166001600160a01b039091161782556020909201916001909101906200028d565b50620002d292915062000322565b5090565b82805482825590600052602060002090810192821562000314579160200282015b8281111562000314578251825591602001919060010190620002f7565b50620002d29291506200034c565b6200034991905b80821115620002d25780546001600160a01b031916815560010162000329565b90565b6200034991905b80821115620002d2576000815560010162000353565b61090a80620003796000396000f3fe608060405234801561001057600080fd5b50600436106100885760003560e01c8063949d225d1161005b578063949d225d146101be5780639da0fd1c146101c6578063aa6ca808146101ce578063b6839baf146101d657610088565b806314e9fb011461008d57806345b4567e146100bc5780634f64b2be14610145578063868bc83d1461018b575b600080fd5b6100aa600480360360208110156100a357600080fd5b50356101f3565b60408051918252519081900360200190f35b6100f5600480360360408110156100d257600080fd5b508035906020013573ffffffffffffffffffffffffffffffffffffffff16610211565b60408051602080825283518183015283519192839290830191858101910280838360005b83811015610131578181015183820152602001610119565b505050509050019250505060405180910390f35b6101626004803603602081101561015b57600080fd5b5035610539565b6040805173ffffffffffffffffffffffffffffffffffffffff9092168252519081900360200190f35b6100aa600480360360208110156101a157600080fd5b503573ffffffffffffffffffffffffffffffffffffffff1661056d565b6100aa61057f565b6100aa610585565b6100f561058b565b6100f5600480360360208110156101ec57600080fd5b50356105fa565b6002818154811061020057fe5b600091825260209091200154905081565b606080600354604051908082528060200260200182016040528015610240578160200160208202803883390190505b50905060005b6003548110156104ca57600460006001838154811061026157fe5b9060005260206000200160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020548473ffffffffffffffffffffffffffffffffffffffff1663868bc83d600184815481106102f057fe5b600091825260209182902001546040805163ffffffff851660e01b815273ffffffffffffffffffffffffffffffffffffffff90921660048301525160248083019392829003018186803b15801561034657600080fd5b505afa15801561035a573d6000803e3d6000fd5b505050506040513d602081101561037057600080fd5b505111156104c2576104a9600460006001848154811061038c57fe5b9060005260206000200160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020548573ffffffffffffffffffffffffffffffffffffffff1663868bc83d6001858154811061041b57fe5b600091825260209182902001546040805163ffffffff851660e01b815273ffffffffffffffffffffffffffffffffffffffff90921660048301525160248083019392829003018186803b15801561047157600080fd5b505afa158015610485573d6000803e3d6000fd5b505050506040513d602081101561049b57600080fd5b50519063ffffffff61068f16565b8282815181106104b557fe5b6020026020010181815250505b600101610246565b5060005b60035481101561052f576105106000546105048484815181106104ed57fe5b6020026020010151886106d890919063ffffffff16565b9063ffffffff61074b16565b82828151811061051c57fe5b60209081029190910101526001016104ce565b5090505b92915050565b6001818154811061054657fe5b60009182526020909120015473ffffffffffffffffffffffffffffffffffffffff16905081565b60046020526000908152604090205481565b60035481565b60005481565b606060018054806020026020016040519081016040528092919081815260200182805480156105f057602002820191906000526020600020905b815473ffffffffffffffffffffffffffffffffffffffff1681526001909101906020018083116105c5575b5050505050905090565b606080600354604051908082528060200260200182016040528015610629578160200160208202803883390190505b50905060005b600354811015610688576106696000546105046002848154811061064f57fe5b9060005260206000200154876106d890919063ffffffff16565b82828151811061067557fe5b602090810291909101015260010161062f565b5092915050565b60006106d183836040518060400160405280601e81526020017f536166654d6174683a207375627472616374696f6e206f766572666c6f77000081525061078d565b9392505050565b6000826106e757506000610533565b828202828482816106f457fe5b04146106d1576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825260218152602001806108be6021913960400191505060405180910390fd5b60006106d183836040518060400160405280601a81526020017f536166654d6174683a206469766973696f6e206279207a65726f00000000000081525061083e565b60008184841115610836576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004018080602001828103825283818151815260200191508051906020019080838360005b838110156107fb5781810151838201526020016107e3565b50505050905090810190601f1680156108285780820380516001836020036101000a031916815260200191505b509250505060405180910390fd5b505050900390565b600081836108a7576040517f08c379a00000000000000000000000000000000000000000000000000000000081526020600482018181528351602484015283519092839260449091019190850190808383600083156107fb5781810151838201526020016107e3565b5060008385816108b357fe5b049594505050505056fe536166654d6174683a206d756c7469706c69636174696f6e206f766572666c6f77a165627a7a723058201e5e31c6d769eb32e3a62da63aa438ca035a93b215f1cd717d914e1b5af0973e0029`

// DeployBasket deploys a new Ethereum contract, binding an instance of Basket to it.
func DeployBasket(auth *bind.TransactOpts, backend bind.ContractBackend, _tokens []common.Address, _backing []*big.Int, _frontTokenDecimals *big.Int) (common.Address, *types.Transaction, *Basket, error) {
	parsed, err := abi.JSON(strings.NewReader(BasketABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(BasketBin), backend, _tokens, _backing, _frontTokenDecimals)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Basket{BasketCaller: BasketCaller{contract: contract}, BasketTransactor: BasketTransactor{contract: contract}, BasketFilterer: BasketFilterer{contract: contract}}, nil
}

// Basket is an auto generated Go binding around an Ethereum contract.
type Basket struct {
	BasketCaller     // Read-only binding to the contract
	BasketTransactor // Write-only binding to the contract
	BasketFilterer   // Log filterer for contract events
}

// BasketCaller is an auto generated read-only Go binding around an Ethereum contract.
type BasketCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BasketTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BasketTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BasketFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BasketFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BasketSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BasketSession struct {
	Contract     *Basket           // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BasketCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BasketCallerSession struct {
	Contract *BasketCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// BasketTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BasketTransactorSession struct {
	Contract     *BasketTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// BasketRaw is an auto generated low-level Go binding around an Ethereum contract.
type BasketRaw struct {
	Contract *Basket // Generic contract binding to access the raw methods on
}

// BasketCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BasketCallerRaw struct {
	Contract *BasketCaller // Generic read-only contract binding to access the raw methods on
}

// BasketTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BasketTransactorRaw struct {
	Contract *BasketTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBasket creates a new instance of Basket, bound to a specific deployed contract.
func NewBasket(address common.Address, backend bind.ContractBackend) (*Basket, error) {
	contract, err := bindBasket(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Basket{BasketCaller: BasketCaller{contract: contract}, BasketTransactor: BasketTransactor{contract: contract}, BasketFilterer: BasketFilterer{contract: contract}}, nil
}

// NewBasketCaller creates a new read-only instance of Basket, bound to a specific deployed contract.
func NewBasketCaller(address common.Address, caller bind.ContractCaller) (*BasketCaller, error) {
	contract, err := bindBasket(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BasketCaller{contract: contract}, nil
}

// NewBasketTransactor creates a new write-only instance of Basket, bound to a specific deployed contract.
func NewBasketTransactor(address common.Address, transactor bind.ContractTransactor) (*BasketTransactor, error) {
	contract, err := bindBasket(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BasketTransactor{contract: contract}, nil
}

// NewBasketFilterer creates a new log filterer instance of Basket, bound to a specific deployed contract.
func NewBasketFilterer(address common.Address, filterer bind.ContractFilterer) (*BasketFilterer, error) {
	contract, err := bindBasket(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BasketFilterer{contract: contract}, nil
}

// bindBasket binds a generic wrapper to an already deployed contract.
func bindBasket(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(BasketABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Basket *BasketRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Basket.Contract.BasketCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Basket *BasketRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Basket.Contract.BasketTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Basket *BasketRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Basket.Contract.BasketTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Basket *BasketCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Basket.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Basket *BasketTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Basket.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Basket *BasketTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Basket.Contract.contract.Transact(opts, method, params...)
}

// Backing is a free data retrieval call binding the contract method 0x14e9fb01.
//
// Solidity: function backing(uint256 ) constant returns(uint256)
func (_Basket *BasketCaller) Backing(opts *bind.CallOpts, arg0 *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Basket.contract.Call(opts, out, "backing", arg0)
	return *ret0, err
}

// Backing is a free data retrieval call binding the contract method 0x14e9fb01.
//
// Solidity: function backing(uint256 ) constant returns(uint256)
func (_Basket *BasketSession) Backing(arg0 *big.Int) (*big.Int, error) {
	return _Basket.Contract.Backing(&_Basket.CallOpts, arg0)
}

// Backing is a free data retrieval call binding the contract method 0x14e9fb01.
//
// Solidity: function backing(uint256 ) constant returns(uint256)
func (_Basket *BasketCallerSession) Backing(arg0 *big.Int) (*big.Int, error) {
	return _Basket.Contract.Backing(&_Basket.CallOpts, arg0)
}

// BackingMap is a free data retrieval call binding the contract method 0x868bc83d.
//
// Solidity: function backingMap(address ) constant returns(uint256)
func (_Basket *BasketCaller) BackingMap(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Basket.contract.Call(opts, out, "backingMap", arg0)
	return *ret0, err
}

// BackingMap is a free data retrieval call binding the contract method 0x868bc83d.
//
// Solidity: function backingMap(address ) constant returns(uint256)
func (_Basket *BasketSession) BackingMap(arg0 common.Address) (*big.Int, error) {
	return _Basket.Contract.BackingMap(&_Basket.CallOpts, arg0)
}

// BackingMap is a free data retrieval call binding the contract method 0x868bc83d.
//
// Solidity: function backingMap(address ) constant returns(uint256)
func (_Basket *BasketCallerSession) BackingMap(arg0 common.Address) (*big.Int, error) {
	return _Basket.Contract.BackingMap(&_Basket.CallOpts, arg0)
}

// FrontTokenDecimals is a free data retrieval call binding the contract method 0x9da0fd1c.
//
// Solidity: function frontTokenDecimals() constant returns(uint256)
func (_Basket *BasketCaller) FrontTokenDecimals(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Basket.contract.Call(opts, out, "frontTokenDecimals")
	return *ret0, err
}

// FrontTokenDecimals is a free data retrieval call binding the contract method 0x9da0fd1c.
//
// Solidity: function frontTokenDecimals() constant returns(uint256)
func (_Basket *BasketSession) FrontTokenDecimals() (*big.Int, error) {
	return _Basket.Contract.FrontTokenDecimals(&_Basket.CallOpts)
}

// FrontTokenDecimals is a free data retrieval call binding the contract method 0x9da0fd1c.
//
// Solidity: function frontTokenDecimals() constant returns(uint256)
func (_Basket *BasketCallerSession) FrontTokenDecimals() (*big.Int, error) {
	return _Basket.Contract.FrontTokenDecimals(&_Basket.CallOpts)
}

// GetTokens is a free data retrieval call binding the contract method 0xaa6ca808.
//
// Solidity: function getTokens() constant returns(address[])
func (_Basket *BasketCaller) GetTokens(opts *bind.CallOpts) ([]common.Address, error) {
	var (
		ret0 = new([]common.Address)
	)
	out := ret0
	err := _Basket.contract.Call(opts, out, "getTokens")
	return *ret0, err
}

// GetTokens is a free data retrieval call binding the contract method 0xaa6ca808.
//
// Solidity: function getTokens() constant returns(address[])
func (_Basket *BasketSession) GetTokens() ([]common.Address, error) {
	return _Basket.Contract.GetTokens(&_Basket.CallOpts)
}

// GetTokens is a free data retrieval call binding the contract method 0xaa6ca808.
//
// Solidity: function getTokens() constant returns(address[])
func (_Basket *BasketCallerSession) GetTokens() ([]common.Address, error) {
	return _Basket.Contract.GetTokens(&_Basket.CallOpts)
}

// NewQuantitiesRequired is a free data retrieval call binding the contract method 0x45b4567e.
//
// Solidity: function newQuantitiesRequired(uint256 _frontTokenSupply, address _other) constant returns(uint256[])
func (_Basket *BasketCaller) NewQuantitiesRequired(opts *bind.CallOpts, _frontTokenSupply *big.Int, _other common.Address) ([]*big.Int, error) {
	var (
		ret0 = new([]*big.Int)
	)
	out := ret0
	err := _Basket.contract.Call(opts, out, "newQuantitiesRequired", _frontTokenSupply, _other)
	return *ret0, err
}

// NewQuantitiesRequired is a free data retrieval call binding the contract method 0x45b4567e.
//
// Solidity: function newQuantitiesRequired(uint256 _frontTokenSupply, address _other) constant returns(uint256[])
func (_Basket *BasketSession) NewQuantitiesRequired(_frontTokenSupply *big.Int, _other common.Address) ([]*big.Int, error) {
	return _Basket.Contract.NewQuantitiesRequired(&_Basket.CallOpts, _frontTokenSupply, _other)
}

// NewQuantitiesRequired is a free data retrieval call binding the contract method 0x45b4567e.
//
// Solidity: function newQuantitiesRequired(uint256 _frontTokenSupply, address _other) constant returns(uint256[])
func (_Basket *BasketCallerSession) NewQuantitiesRequired(_frontTokenSupply *big.Int, _other common.Address) ([]*big.Int, error) {
	return _Basket.Contract.NewQuantitiesRequired(&_Basket.CallOpts, _frontTokenSupply, _other)
}

// QuantitiesRequired is a free data retrieval call binding the contract method 0xb6839baf.
//
// Solidity: function quantitiesRequired(uint256 _frontTokenSupply) constant returns(uint256[])
func (_Basket *BasketCaller) QuantitiesRequired(opts *bind.CallOpts, _frontTokenSupply *big.Int) ([]*big.Int, error) {
	var (
		ret0 = new([]*big.Int)
	)
	out := ret0
	err := _Basket.contract.Call(opts, out, "quantitiesRequired", _frontTokenSupply)
	return *ret0, err
}

// QuantitiesRequired is a free data retrieval call binding the contract method 0xb6839baf.
//
// Solidity: function quantitiesRequired(uint256 _frontTokenSupply) constant returns(uint256[])
func (_Basket *BasketSession) QuantitiesRequired(_frontTokenSupply *big.Int) ([]*big.Int, error) {
	return _Basket.Contract.QuantitiesRequired(&_Basket.CallOpts, _frontTokenSupply)
}

// QuantitiesRequired is a free data retrieval call binding the contract method 0xb6839baf.
//
// Solidity: function quantitiesRequired(uint256 _frontTokenSupply) constant returns(uint256[])
func (_Basket *BasketCallerSession) QuantitiesRequired(_frontTokenSupply *big.Int) ([]*big.Int, error) {
	return _Basket.Contract.QuantitiesRequired(&_Basket.CallOpts, _frontTokenSupply)
}

// Size is a free data retrieval call binding the contract method 0x949d225d.
//
// Solidity: function size() constant returns(uint256)
func (_Basket *BasketCaller) Size(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Basket.contract.Call(opts, out, "size")
	return *ret0, err
}

// Size is a free data retrieval call binding the contract method 0x949d225d.
//
// Solidity: function size() constant returns(uint256)
func (_Basket *BasketSession) Size() (*big.Int, error) {
	return _Basket.Contract.Size(&_Basket.CallOpts)
}

// Size is a free data retrieval call binding the contract method 0x949d225d.
//
// Solidity: function size() constant returns(uint256)
func (_Basket *BasketCallerSession) Size() (*big.Int, error) {
	return _Basket.Contract.Size(&_Basket.CallOpts)
}

// Tokens is a free data retrieval call binding the contract method 0x4f64b2be.
//
// Solidity: function tokens(uint256 ) constant returns(address)
func (_Basket *BasketCaller) Tokens(opts *bind.CallOpts, arg0 *big.Int) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Basket.contract.Call(opts, out, "tokens", arg0)
	return *ret0, err
}

// Tokens is a free data retrieval call binding the contract method 0x4f64b2be.
//
// Solidity: function tokens(uint256 ) constant returns(address)
func (_Basket *BasketSession) Tokens(arg0 *big.Int) (common.Address, error) {
	return _Basket.Contract.Tokens(&_Basket.CallOpts, arg0)
}

// Tokens is a free data retrieval call binding the contract method 0x4f64b2be.
//
// Solidity: function tokens(uint256 ) constant returns(address)
func (_Basket *BasketCallerSession) Tokens(arg0 *big.Int) (common.Address, error) {
	return _Basket.Contract.Tokens(&_Basket.CallOpts, arg0)
}
