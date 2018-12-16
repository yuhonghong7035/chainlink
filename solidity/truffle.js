require('babel-register');
require('babel-polyfill');

module.exports = {
  network: "development",
  compilers: {
    solc: {
      version: "0.4.24",
      docker: true,
    }
  },
  networks: {
    development: {
      host: "127.0.0.1",
      port: 8545,
      network_id: "*",
      gas: 4700000
    }
  }
};
