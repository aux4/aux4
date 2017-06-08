#!/usr/bin/env node

const config = require('./lib/config');
const executor = require('./lib/executor');

config.load(function(){
  executor.init();
  executor.execute(process.argv.splice(2));
});
