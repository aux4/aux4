const logExecutor = require('./executors/logExecutor');
const profileExecutor = require('./executors/profileExecutor');
const commandLineExecutor = require('./executors/commandLineExecutor');

const executors = [
  logExecutor, commandLineExecutor
];

module.exports = {
  execute: function(commands, args) {
    commands.forEach(function(command){
      for (let i = 0; i < executors.length; i++) {
        let executor = executors[i];
        if (executor.execute(command, args) === true) {
          break;
        }
      }
    });
  }
};
