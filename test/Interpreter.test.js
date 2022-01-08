const parameterInterpreter = require("../lib/interpreters/ParameterInterpreter");

const Interpreter = require("../lib/Interpreter");
const interpreter = new Interpreter();

describe("interpreter", () => {
  describe("interpret", () => {
    let result, command, action, args, parameters;

    beforeEach(() => {
      jest.resetModules();
      parameterInterpreter.interpret = jest.fn();
      interpreter.add(parameterInterpreter);

      command = {};
      args = [];
      parameters = {};
      action = "";

      result = interpreter.interpret(command, action, args, parameters);
    });

    it("calls parameterInterpreter", () => {
      expect(parameterInterpreter.interpret).toHaveBeenCalledWith(command, action, args, parameters);
    });
  });
});
