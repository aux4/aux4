const colors = require('colors');
const childProcess = require('child_process');

const out = require('../output');
const interpreter = require('../interpreter');

module.exports = {
  execute: function(command, action, args, parameters) {
    let cmd = interpreter.interpret(command, action, args, parameters);

    try {
      let result = childProcess.execSync(cmd);
      out.println(result.toString());
      return true;
    } catch (err) {
      out.println(err.stdout);
      out.println(err.message.red);
      throw err;
    }
  }
};
