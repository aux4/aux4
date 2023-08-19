const fs = require("fs");
const childProcess = require("child_process");

class VersionCommand {
  static execute() {
    const year = new Date().getFullYear();

    const packageJson = JSON.parse(fs.readFileSync(__dirname + "/../../package.json", { encoding: "utf8" }));
    console.log();
    console.log(`  ${"aux4".cyan} ${getFormattedVersion(packageJson.version)}`);
    console.log(`  ${year} Aux4. Aux4 is created and maintained by aux4 community.`.gray);
    console.log(`  https://aux4.io`.gray);
    console.log();

    const output = childProcess.execSync("npm view aux4 --json");
    const response = output.toString().trim();

    const aux4Info = JSON.parse(response);
    const latestVersion = aux4Info["dist-tags"].latest;

    if (latestVersion !== packageJson.version) {
      console.log(`Latest version: ${getFormattedVersion(latestVersion)}`);
      console.log(`Run ${"aux4 upgrade".cyan} to upgrade to the latest version`);
      console.log();
    }
  }
}

function getFormattedVersion(version) {
  return `v${version}`.yellow;
}

module.exports = VersionCommand;
