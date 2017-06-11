#!/usr/bin/env node

const config = require('./lib/config');
const executor = require('./lib/executor');
const executorChain = require('./lib/executorChain');

const logExecutor = require('./lib/executors/logExecutor');
const profileExecutor = require('./lib/executors/profileExecutor');
const commandLineExecutor = require('./lib/executors/commandLineExecutor');

executorChain.add(logExecutor);
executorChain.add(profileExecutor);
executorChain.add(commandLineExecutor);

config.load(function(){
  executor.init();
  executor.execute(process.argv.splice(2));
});
