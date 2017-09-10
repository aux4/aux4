const colors = require('colors');
const childProcess = require('child_process');

const out = require('../output');
const interpreter = require('../interpreter');
const JSON_PREFIX = 'json:';

module.exports = {
  execute: function(command, action, args, parameters) {
    let cmd = interpreter.interpret(command, action, args, parameters);

    try {
      let output = childProcess.execSync(cmd.replace(JSON_PREFIX, ''));
      let response = output.toString();
      if (action.startsWith(JSON_PREFIX)) {
        response = JSON.parse(response);
      }
      parameters['response'] = response;
      out.println(output.toString());
      return true;
    } catch (err) {
      out.println(err.stdout);
      out.println(err.message.red);
      throw err;
    }
  }
};
