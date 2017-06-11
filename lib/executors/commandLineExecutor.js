const childProcess = require('child_process');

const out = require('../output');
const interpreter = require('../interpreter');

module.exports = {
  execute: function(command, action, args, parameters) {
    let cmd = interpreter.interpret(action, args, parameters);
    childProcess.exec(cmd, function(err, stdout, stderr) {
      if (err) {
        out.println(stderr);
        return;
      }
      out.println(stdout);
    });
    return true;
  }
};
