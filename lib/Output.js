const Output = {
  print: function () {
    const text = formatOutputText(arguments);
    process.stdout.write(text);
  },
  println: function () {
    const text = formatOutputText(arguments) + "\n";
    process.stdout.write(text);
  }
};

function formatOutputText(args) {
  let text = "";
  for (let i = 0; i < args.length; i++) {
    const arg = args[i];
    if (arg === undefined || arg === "") {
      continue;
    }
    if (i > 0 && text.length > 0) {
      text += " ";
    }
    text += arg;
  }
  return text;
}

module.exports = Output;
