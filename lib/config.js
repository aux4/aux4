const colors = require('colors');
const fs = require('fs');

const out = require('./output');

const CONFIG_FILE_NAME = '.aux4';

let config = {profiles: []};

module.exports = {
  file: function() {
    return config;
  },
  load: function() {
    fs.access(CONFIG_FILE_NAME, function(err) {
      if (err) {
        out.println(`${CONFIG_FILE_NAME} file not found`.red);
        return;
      }

      fs.readFile(CONFIG_FILE_NAME, function(err, data) {
        if (err) {
          out.println('error reading .aux file, check the permissions'.red);
          return;
        }

        try {
          config = JSON.parse(data);
        } catch (e) {
          out.println('.aux4 is not a valid json file'.red);
        }
      });
    });
  }
};
