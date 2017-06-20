const colors = require('colors');
const fs = require('fs');

const out = require('./output');

const CONFIG_FILE_NAME = '.aux4';

const profile$ = {
  name: '$',
  commands: [
    {
      value: 'encrypt',
      execute: [
        'crypto:encrypt'
      ]
    }
  ]
};

let config = {profiles: []};

module.exports = {
  $: function() {
    return profile$;
  },
  file: function() {
    return config;
  },
  load: function(callback) {
    fs.access(CONFIG_FILE_NAME, function(err) {
      if (err) {
        out.println(`${CONFIG_FILE_NAME} file not found`.red);
        callback(new Error(`${CONFIG_FILE_NAME} file not found`));
        return;
      }

      fs.readFile(CONFIG_FILE_NAME, function(err, data) {
        if (err) {
          out.println(`error reading ${CONFIG_FILE_NAME} file, check the permissions`.red);
          callback(new Error(`error reading ${CONFIG_FILE_NAME} file, check the permissions`));
          return;
        }

        try {
          config = JSON.parse(data);
        } catch (e) {
          out.println(`${CONFIG_FILE_NAME} is not a valid json file`.red);
          callback(new Error(`${CONFIG_FILE_NAME} is not a valid json file`));
        }

        callback(undefined);
      });
    });
  }
};
