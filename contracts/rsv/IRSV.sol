pragma solidity ^0.5.8;

interface IRSV {
    // Standard ERC20 functions
    function transfer(address, uint256) external returns(bool);
    function approve(address, uint256) external returns(bool);
    function transferFrom(address, address, uint256) external returns(bool);
    function totalSupply() external view returns(uint256);
    function balanceOf(address) external view returns(uint256);
    function allowance(address, address) external view returns(uint256);
    event Transfer(address indexed from, address indexed to, uint256 value);
    event Approval(address indexed owner, address indexed spender, uint256 value);

    // RSV-specific functions
    function decimals() external view returns(uint8);
    function mint(address, uint256) external;
    function burnFrom(address, uint256) external;
}
