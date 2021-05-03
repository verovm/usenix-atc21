const fs = require("fs");
const readline = require('readline');
const Web3 = require('web3');
const web3 = new Web3(new Web3.providers.HttpProvider("ws://localhost:8546"));
const Contract = require('web3-eth-contract');
const version = web3.version.api;
const request = require('request');
const async = require('async');
const Set = require("collections/set");

const LAST_LINE_FILE = "last_line.txt"
const CONTRACTS = new Set();
const ADDR_MAP = new Set();

// Start processing from the last line, at the end of processing store the last line back in the file
let currentLines = 0;
let lastLine = 0;
if (fs.existsSync(LAST_LINE_FILE)) lastLine = parseInt(fs.readFileSync(LAST_LINE_FILE));


/**
 * Read data about blocks and trigger Trie analysis
 * @param file
 * @param cb callback
 * @param onDone invoked when file is processed
 */
function readStorageData(file, cb, saveData, onDone) {

    const accounts = fs.createReadStream(file);
    const rl = readline.createInterface({
        input: accounts,
        crlfDelay: Infinity
    });

    // we run all IOs in parallel as ETHScan allows only 5 requests/s
    let tasks = []
    rl.on('line', line => {
        const items = line.split(",");
        const blockNumber = items[0];
        const accountAddress = items[1];
        const storageRoot = items[2];

        tasks.push(function (done) {
            setTimeout(()=>cb(accountAddress, saveData, done), 20);    // only 5 requests per sec can be send.  Wait rawly 200ms * 5 = 1s
        })
    });

    // Task do run everything in series
    const triggerRequests = function() {
        async.series(tasks, onDone);
    }

    rl.on('close', triggerRequests);
}


/**
 * Read addresses from name of the directories
 * @param file
 */
function readAddressesFromDirs(file, cb, saveData, onDone) {

    const dir = fs.opendirSync(file)
    let tasks = []
    let dirent;

    console.log("DEBUG: START: From line: " + lastLine);

    while ((dirent = dir.readSync()) !== null) {
        if (currentLines >= lastLine) {
            const address = dirent.name;
            tasks.push(function (done) {
                setTimeout(()=> cb(address, saveData, done), 1000);    // one second delay - but we trigger 5 request in parallel
            })
        } else currentLines++;  // skipping already processed lines
    }
    dir.closeSync()

    console.log("DEBUG: TASKs submitted: " + tasks.length);

    // we may send 5 request/s -
    async.parallelLimit(tasks, 5, onDone);
}

/**
 * Download SmartContract from ethscan
 * @param blockNumber
 * @param accountAddress
 * @param storageRoot
 */
function downloadContract(accountAddress, onData, onDone) {

    const addr = accountAddress
    const apiKey = "94KHWEDZWY7GMKADY72GDTD4BQUX4K82E3"
    const options = {json: true};

    // const taskAbi = function (done) {
    //     const abiUrl = "https://api.etherscan.io/api?module=contract&action=getabi&address=" + addr + "&apikey=" + apiKey;
    //     request(abiUrl, options, (error, res, body) => {
    //         if (!error && res.statusCode === 200) {
    //             const abi = JSON.parse(body.result);
    //             done(null, body.result);
    //         } else {
    //             console.log("Error: " + error + " Status code: " + res.statusCode);
    //             done(error);
    //         }
    //     });
    // }

    const taskSrc = function (done) {
        const srcUrl = "https://api.etherscan.io/api?module=contract&action=getsourcecode&address=" + addr + "&apikey=" + apiKey
        const URL_ID = Math.floor(Math.random() * Math.floor(100000));
        console.log("DEBUG: API URL_ID: " + URL_ID + ", request: " + srcUrl);
        request(srcUrl, options, (error, res, body) => {
            if (!error && res.statusCode === 200 && body.status === "1") {
                const contrName = body.result[0].ContractName
                const abi = body.result[0].ABI
                const src = body.result[0].SourceCode
                // const json = JSON.parse(body.result);
                done(null, {contactName: contrName, abi: abi, src : src, url_id: URL_ID} );
            } else {
                if (error) {
                    console.log("Error: " + error);
                    done(error);
                } else
                if (res.statusCode === 200 && body.status !== "1") {
                    console.log("Error: " + error + " Status code: " + res.statusCode + " Status: " + body.status + " Result: " + body.result);
                    done(body.result);
                } else {
                    console.log("Error status: " + res.statusCode);
                    done(res.statusCode);
                }
            }
        });
    }

    const onDoneSeries = function (err, results) {
        if (!err)  {
            const contrName = results[0].contactName
            const abi = results[0].abi
            const src = results[0].src
            const url_id = results[0].url_id;

            console.log("DEBUG: URL done: " + url_id)

            // !!! Contract fuzzer ignores situation when there is a contract with the same name under more addresses - so we store just the names as well
            if (CONTRACTS.size >= MAX_RESULTS) {
                console.log("DEBUG: Contract batch done: " + CONTRACTS.size);
                onDone("OK");  // ignore all results above -- send fake error not to continue
                console.log("DEBUG: terminating: " + CONTRACTS.size);
                return
            }

            if (contrName !== undefined && abi !== "Contract source code not verified") {
                onData(addr, contrName, abi, src)
                console.log("DEBUG: Contract DOES have a verified source code,  NEW #contracts: " + CONTRACTS.size);

                if (CONTRACTS.size % 100 === 0) console.log("Processed unique contract names: " + CONTRACTS.size)
            }

            // console.log("res" + results)
        }

        currentLines++; // increase processed lines - does not matter if successful or not
        // never report error - we want to run all API requests, even if some of them fails
        onDone()
    }

    // async.series([taskAbi, taskSrc], onDone);
    async.series([taskSrc], onDoneSeries);
}


/** MAIN */

const args = process.argv.slice(2);
const inputDir = args[0];
const outputDir = args[1];
const MAX_RESULTS = args[2];

// init directories
fs.mkdirSync(outputDir + "/fuzzer/config/", { recursive: true });
fs.mkdirSync(outputDir + "/fuzzer/reporter/bug/", { recursive: true });
fs.mkdirSync(outputDir + "/verified_contract_abis/", { recursive: true });

// const CSV_FILE = "accounts_storage_11000000.csv"

const storeContractData = function (addr, name, abi, bytes) {
    // save abi into a file
    const abiStream = fs.createWriteStream(outputDir + "/verified_contract_abis/" + name + ".abi");
    abiStream.write(abi, ()=> abiStream.end());

    // write addrMap
    const newLine = [];
    newLine.push(addr);
    newLine.push(name);

    ADDR_MAP.add(newLine.join(','));
    CONTRACTS.add(name);
}

const writeResults = function () {
    console.log("DEBUG: END: Last line: " + currentLines);

    const addrMapStream = fs.createWriteStream(outputDir + "/fuzzer/config/addrmap.csv");
    const contractListStream = fs.createWriteStream(outputDir + "/fuzzer/config/contracts.list");

    let tasks1 = []
    for (let contract of CONTRACTS) tasks1.push( done => contractListStream.write(contract + '\n', done));
    let tasks2 = []
    for (let addr of ADDR_MAP) tasks2.push( done => addrMapStream.write(addr + '\n', done));

    // make sure to close files when all is written
    const end1 = done => async.series(tasks1, () => contractListStream.end(done))
    const end2 = done => async.series(tasks2, () => addrMapStream.end(done))

    const end3 = done => fs.writeFile(LAST_LINE_FILE, (currentLines).toString(), done);

    console.log("DEBUG: all written " + currentLines);

    // TODO - hard exit - it was sometimes freezing, not sure where/why
    // make sure to hard exit when all streams are finished
    async.parallel([end1, end2, end3], ()=>process.exit())
}


// readStorageData(CSV_FILE, downloadContract, saveData, closeFiles)

// downloadContract( "0x5dcd7caf96c3c563cb4ca2628e9a29a99097d545", (addr, name, abi, bytes) => {
//     saveData(addr, name, abi, bytes);
//     closeFiles();
// });
// downloadContract( "0xfb6916095ca1df60bb79ce92ce3ea74c37c5d359", (addr, name, abi, bytes) => {
//     saveData(addr, name, abi, bytes);
//     closeFiles();
// });

readAddressesFromDirs(inputDir, downloadContract, storeContractData, writeResults);


