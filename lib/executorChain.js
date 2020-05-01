const colors = require('colors');
const out = require('./output');

const executors = [];

module.exports = {
  add: function(executor) {
    executors.push(executor);
  },
  execute: function(command, args, parameters) {
    const actions = command.execute;
    if (actions === undefined) {
      out.println('execute is not defined'.red);
      return;
    }

    for (let x = 0; x < actions.length; x++) {
      const action = actions[x];
      for (let i = 0; i < executors.length; i++) {
        const executor = executors[i];
        let response;

        try {
          response = executor.execute(command, action, args, parameters);
        } catch (e) {
          return;
        }

        if (response) {
          break;
        }
      }
    }
  }
};
