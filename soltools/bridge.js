// This is the JavaScript end of the Go-JavaScript bridge implemented in this package.

// Imports
const util = require('util');
const http = require('http');
const Web3 = require('web3');

const { SolCompilerArtifactAdapter } = require('@0x/sol-trace');
const { CoverageSubprovider } = require('@0x/sol-coverage');
const ProviderEngine = require('web3-provider-engine');
const RpcSubprovider = require('web3-provider-engine/subproviders/rpc.js');

// Command-line arguments.
const [, , artifactsDir, contractsDir] = process.argv;

// Create web3 provider chain.
// We need an artifact adapter so the coverage subprovider knows how to map EVM traces source code.
// We need a coverage subprovider so we can write a coverage report.
// We also need an RPC subprovider to handle everything else.
const artifactAdapter = new SolCompilerArtifactAdapter(artifactsDir, contractsDir);
const coverageSubprovider = new CoverageSubprovider(
  artifactAdapter,
  '0x5409ed021d9299bf6814279a6a1411a7e866a631'
);
const provider = new ProviderEngine();
provider.addProvider(coverageSubprovider);
provider.addProvider(new RpcSubprovider({rpcUrl: 'http://localhost:8545'}));
provider.start();
provider.stop();
provider.send = provider.sendAsync.bind(provider);
const web3 = new Web3(provider);

// readFull returns a Promise that resolves to the contents of the entire stream.
function readFull(stream) {
  return new Promise((resolve, reject) => {
    var content = '';
    stream.on('data', (buf) => { content += buf.toString(); });
    stream.on('end', () => { resolve(content) });
  });
}

// promisify is a simple Promisify implementation.
function promisify(f) {
  return new Promise((resolve, reject) => f((err, value) => {
    err ? reject(err) : resolve(value);
  }));
}

// rpcs contains implementations of all of the RPCs we support, keyed by
// their "method name".
const rpcs = {
  // RPCs that wrap web3 calls.
  pendingNonceAt: address => promisify(cb => web3.eth.getTransactionCount(address, 'pending', cb)),
  sendTransaction: tx => promisify(cb => web3.eth.sendRawTransaction(tx, cb)),
  estimateGas: callObject => promisify(cb => web3.eth.estimateGas(callObject, cb)),
  call: ({call, block}) => promisify(cb => web3.eth.call(call, block, cb)),

  // Other RPCs.
  writeCoverage: _ => coverageSubprovider.writeCoverageAsync().then(_ => true),
	close: _ => {
    setImmediate(server.close.bind(server), (err, value) => {
      provider.stop();
      console.log('javascript web3 server stopped');
    })
		return true;
	},
};

// Create the web server and listen for requests.
const port = 3000;
const server = http.createServer(async (request, response) => {
  try {
    // Simple RPC endpoint: requests are JSON-formatted as {"method": <method>, "data": <arbitrary JSON argument for method>}.
    const content = await readFull(request);
    const {method, data} = JSON.parse(content);
    // Our RPCs are synchronous, so await the results.
    response.end(JSON.stringify(await rpcs[method](data)));

  }
  catch(error) {
    response.statusCode = 500;
    response.end(JSON.stringify(error));
  }
});
server.listen(port, (err) => {
  if (err) {
    return console.log('server error', err)
  }
  console.log(`javascript web3 server listening on port ${port}`)
});
