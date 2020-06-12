const fs = require('fs');
const homePath = require('os').homedir();

const colors = require('colors');

const out = require('../output');

const AUX4_PACKAGE_DIRECTORY = '/.aux4.config/packages/';

const LIST_PREFIX = 'package:list';
const INSTALL_PREFIX = 'package:install';
const UNINSTALL_PREFIX = 'package:uninstall';

module.exports = {
  execute: function(command, action, args, parameters) {
    if (action.startsWith(LIST_PREFIX)) {
      list(args, parameters);
      return true;
    } else if (action.startsWith(INSTALL_PREFIX)) {
      install(args, parameters);
      return true;
    } else if (action.startsWith(UNINSTALL_PREFIX)) {
      uninstall(args, parameters);
      return true;
    } else {
      return false;
    }
  }
};

function list() {
  if (fs.existsSync(homePath + AUX4_PACKAGE_DIRECTORY)) {
    fs.readdir(homePath + AUX4_PACKAGE_DIRECTORY, (err, files) => {
      files.forEach(file => {
        out.println('- ' + file.replace('.json', '').yellow);
      });
    });
  }
}

function install(args, parameters) {
  const thePackage = args[0];

  if (!thePackage) {
    out.println('You must specify the package file'.red);
    return;
  }

  if (!fs.existsSync(homePath + AUX4_PACKAGE_DIRECTORY)) {
    fs.mkdirSync(homePath + AUX4_PACKAGE_DIRECTORY, {recursive: true});
  }

  if (fs.existsSync(thePackage)) {
    let aux4File;

    try {
      aux4File = fs.readFileSync(thePackage).toJSON();
    } catch (e) {
      out.println(`The package ${thePackage} is not a valid json file`.red);
      return;
    }

    if (aux4File.package === undefined) {
      out.println(`There is no package information in the JSON file`.red);
      return;
    }

    if (!aux4File.package.name) {
      out.println(`There is no package name in the JSON file`.red);
      return;
    }

    if (!aux4File.package.version) {
      out.println(`There is no package version in the JSON file`.red);
      return;
    }

    const packageName = aux4File.package.name;
    const packageVersion = aux4File.package.version;
    const packageFileName = `${packageName}.json`;

    out.println(`Installing ${thePackage} as ${packageName} version ${packageVersion}`);

    fs.writeFileSync(homePath + AUX4_PACKAGE_DIRECTORY + packageFileName, JSON.stringify(aux4File));

    out.println(`Package ${packageName} was installed`);
  } else {
    out.println(`Package ${thePackage} file not found`.red);
  }
}

function uninstall(args, parameters) {
  const thePackage = parameters.name || args[0];

  if (!thePackage) {
    out.println('You must specify the package to be uninstalled'.red);
    return;
  }

  const path = homePath + AUX4_PACKAGE_DIRECTORY + thePackage + '.json';
  if (!fs.existsSync(path)) {
    out.println(`Package "${thePackage}" does not exist`.red);
    return;
  }

  out.println(`Uninstalling ${thePackage}`);
  fs.unlinkSync(path);

  out.println(`Package ${thePackage} was uninstalled`);
}
