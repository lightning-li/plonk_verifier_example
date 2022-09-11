import { HardhatUserConfig } from "hardhat/config";
import "@nomicfoundation/hardhat-toolbox";
import "hardhat-gas-reporter"

const config: HardhatUserConfig = {
  solidity: {
    version: "0.6.4",
    settings: {
      optimizer: {
        enabled: true,
        runs: 1000,
      },
    },
  },

  etherscan: {
    // Your API key for Etherscan
    // Obtain one at https://bscscan.com/
    apiKey: {
      bscTestnet: '39VGD16VJN8CAUCCB5W7JJ72DZTCDWJJB9'
    }
  },

  networks: {
    // hardhat: {
    //     allowUnlimitedContractSize: true,
    // },
    local: {
      url: "http://127.0.0.1:8545",
      accounts: ['906d5dc5a8ec5050a21987278d42af90852724df53a576e66057990ee48ac269'],
      timeout: 100000,
    },
    BSCTestnet: {
      url: "https://data-seed-prebsc-1-s1.binance.org:8545",
      accounts: ['acbaa269bd7573ff12361be4b97201aef019776ea13384681d4e5ba6a88367d9'],
      timeout: 300000,
      gas: 15000000
    },
  },
};

export default config;
