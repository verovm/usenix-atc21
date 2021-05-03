/*
    This is a server that orchestrates experiments
 */

const express = require("express");
const fs = require("fs");
const async = require('async');

const app = express();

Statistics = class {

    constructor(task) {
        this.task = task;
        this.time = 0;
        this.messages = 0;
        this.totalTime = 0;
        this.speeds = []
    }

    addValue(time) {
        this.time += time;
        this.messages++;
    }

    addSpeed(speed) {
        this.speeds.push(speed);
    }

    speed() {
        let i = 0;
        let sum = 0;
        this.speeds.forEach((s)=>{
            sum += s;
            i++;
        })

        return sum / i;
    }
}


function readTasks(inputDir) {
    // read all available tasks
    const dir = fs.opendirSync(inputDir)
    let tasks = []
    let dirent;

    while ((dirent = dir.readSync()) !== null) {
        tasks.push(parseInt(dirent.name));
    }

    dir.closeSync();
    return tasks.sort((a, b)=> a-b);
}

function server() {

    const rnd = new Date().toISOString()
        .replace(/:/, '-')
        .replace(/:/, '-')
        .replace(/T/, '_')
        .replace(/\..+/, '')
    //Math.floor(Math.random() * 1000000);
    const tasks = readTasks(inputDir);
    const stat = new Map();
    let startTime = Date.now();
    let finishedTasks = 0;

    const speedWath = function () {
        // dump througput at every dump
        if (finishedTasks > 0) {
            const currentTime = Date.now();
            const diffTime = (currentTime - startTime) / 1000 / 60   // min
            const speedTasks = finishedTasks / diffTime
            finishedTasks = 0;
            startTime = currentTime;
            const file = "./out/speed_" + rnd + ".csv"
            fs.appendFileSync(file, speedTasks+ '\n');
        }
    }
    const speedWatchTimer = function () {
        setTimeout(()=> {
            speedWath();
            speedWatchTimer();
        }, 60 * 1000)
    }

    speedWatchTimer();

    const options = { 'flag': 'a+' }
    const lastTaskIndex = fs.readFileSync("./out/lastTask.csv", options);
    let index = 0;
    // find last index
    if (lastTaskIndex.length > 0)
        for (let i = 0; i < tasks.length; i++) {
            index++;
            if (tasks[i] === parseInt(lastTaskIndex)) break;
        }

    app.listen(9999, () => {
        console.log("Server running on port 9999");
    });

    // Return next available task
    app.get("/task", (req, res, next) => {
        const nextTask = index === tasks.length ? "DONE" : tasks[index++]
        console.log("Next task is " + nextTask + " Index: " + index + "/" + tasks.length)
        res.send(nextTask.toString());
    });

    // Add total execution time to process one fuzzing message
    app.get("/results/:task/:time", (req, res, next) => {
        const task = req.params.task;
        const time = parseInt(req.params.time);

        let item = stat.get(task);
        if (item === undefined) {
            item = new Statistics(task);
            stat.set(task, item);
        }
        item.addValue(time);

        res.sendStatus(200);
    });

    // Add total time to finish the whole task - i.e. a batch of 10 contracts
    app.get("/finish/:task/:time", (req, res, next) => {
        const task = req.params.task;
        const time = parseInt(req.params.time);

        let item = stat.get(task);
        item.totalTime = time;

        finishedTasks += parseInt(task);

        res.sendStatus(200);
    });

    // Add recent message speed to process last batch of messages
    app.get("/msgSpeed/:task/:speed", (req, res, next) => {
        const task = req.params.task;
        const speed = parseFloat(req.params.speed);

        let item = stat.get(task);
        if (item === undefined) {
            item = new Statistics(task);
            stat.set(task, item);
        }
        item.addSpeed(speed);
        res.sendStatus(200);
    });

    // dump results
    app.get("/dump", (req, res, next) => {

        // dump all data in a file
        const file = "./out/results_" + rnd + ".csv"
        const stream = fs.createWriteStream( file);
        let tasks = []
        const keys = [...stat.keys()].sort((a, b) => a -b);
        let lastKey = keys[keys.length - 1];
        for (let key of keys) tasks.push( done => {
            const value = stat.get(key);
            const avrg = value.time / value.messages
            const speed = value.speed();
            stream.write(key + "," + value.time + "," + value.messages + "," + avrg + "," + speed + "," + value.totalTime + '\n', done)
        });
        // make sure to close files when all is written
        async.series(tasks, () => stream.end())
        const opt2 = { 'flag': 'w' }
        fs.writeFileSync("./out/lastTask.csv", lastKey +  '\n', opt2);
        res.sendStatus(200)
    });

}

const args = process.argv.slice(2);
const inputDir = args[0];

server();

