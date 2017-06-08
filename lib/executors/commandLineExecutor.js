const childProcess = require('child_process');

const out = require('../output');
const interpreter = require('../interpreter');

module.exports = {
  execute: function(command, args) {
    let cmd = interpreter.interpret(command, args);
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
