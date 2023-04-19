const fs = require("fs");
const homePath = require("os").homedir();

const colors = require("colors");

const out = require("../Output");

const AUX4_PACKAGE_DIRECTORY = "/.aux4.config/packages/";

const LIST_PREFIX = "package:list";
const INSTALL_PREFIX = "package:install";
const UNINSTALL_PREFIX = "package:uninstall";

function PackageExecutor() {
  return {
    execute: async function (command, action, args, parameters) {
      if (action.startsWith(LIST_PREFIX)) {
        await list(args, parameters);
        return true;
      } else if (action.startsWith(INSTALL_PREFIX)) {
        await install(args, parameters);
        return true;
      } else if (action.startsWith(UNINSTALL_PREFIX)) {
        await uninstall(args, parameters);
        return true;
      } else {
        return false;
      }
    }
  };
}

async function list() {
  if (fs.existsSync(homePath + AUX4_PACKAGE_DIRECTORY)) {
    fs.readdir(homePath + AUX4_PACKAGE_DIRECTORY, (err, files) => {
      files.forEach(file => {
        const packageFile = fs.readFileSync(homePath + AUX4_PACKAGE_DIRECTORY + file).toString();
        const packageJson = JSON.parse(packageFile);
        out.println(`- ${packageJson.package.name.yellow} ${packageJson.package.version.yellow}`);
      });
    });
  }
}

async function install(args, parameters) {
  if (!fs.existsSync(homePath + AUX4_PACKAGE_DIRECTORY)) {
    fs.mkdirSync(homePath + AUX4_PACKAGE_DIRECTORY, { recursive: true });
  }

  let thePackage = args.length > 0 ? args[0] : undefined;

  if (thePackage && !fs.existsSync(thePackage)) {
    out.println(`Package ${thePackage} file not found`.red);
    return;
  }

  let aux4File;
  let aux4Json;

  try {
    if (!thePackage) {
      thePackage = 0;
    }

    aux4File = fs.readFileSync(thePackage).toString();
  } catch (e) {
    out.println(`Cannot read the file ${thePackage}`.red);
    return;
  }

  try {
    aux4Json = JSON.parse(aux4File);
  } catch (e) {
    out.println(`The package ${thePackage} is not a valid json`.red);
    return;
  }

  if (aux4Json.package === undefined) {
    out.println(`There is no package information in the JSON file`.red);
    return;
  }

  if (!aux4Json.package.name) {
    out.println(`There is no package name in the JSON file`.red);
    return;
  }

  if (!aux4Json.package.version) {
    out.println(`There is no package version in the JSON file`.red);
    return;
  }

  const packageName = aux4Json.package.name;
  const packageVersion = aux4Json.package.version;
  const packageFileName = `${packageName}.aux4`;

  if (thePackage === 0) {
    thePackage = "input file";
  }

  out.println(`Installing ${thePackage.cyan} as ${packageName.yellow} version ${packageVersion.yellow}`);

  fs.writeFileSync(homePath + AUX4_PACKAGE_DIRECTORY + packageFileName, aux4File);

  out.println(`Package ${packageName.yellow} version ${packageVersion.yellow} has been installed successfully`);
}

async function uninstall(args, parameters) {
  const thePackage = parameters.name || args[0];

  if (!thePackage) {
    out.println("You must specify the package to be uninstalled".red);
    return;
  }

  const path = homePath + AUX4_PACKAGE_DIRECTORY + thePackage + ".aux4";
  if (!fs.existsSync(path)) {
    out.println(`Package "${thePackage}" does not exist`.red);
    return;
  }

  out.println(`Uninstalling ${thePackage}`);
  fs.unlinkSync(path);

  out.println(`Package ${thePackage} was uninstalled`);
}

module.exports = PackageExecutor;
