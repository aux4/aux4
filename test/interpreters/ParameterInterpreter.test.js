const ParameterInterpreter = require("../../lib/interpreters/ParameterInterpreter");
const parameterInterpreter = new ParameterInterpreter();

describe("parameterInterpreter", () => {
  describe("interpret", () => {
    let result, command, args, parameters;

    beforeEach(() => {
      command = {};
    });

    describe("without variables", () => {
      beforeEach(() => {
        args = [];
        parameters = {};
        result = parameterInterpreter.interpret(command, "mkdir test", args, parameters);
      });

      it("does not replace the text", () => {
        expect(result).toEqual("mkdir test");
      });
    });

    describe("with variable and no parameter", () => {
      beforeEach(() => {
        args = [];
        parameters = {};
        result = parameterInterpreter.interpret(command, "echo ${name}", args, parameters);
      });

      it("does not replace the variable", () => {
        expect(result).toEqual("echo ${name}");
      });
    });

    describe("with variable and parameter", () => {
      beforeEach(() => {
        args = [];
        parameters = { name: "John" };
        result = parameterInterpreter.interpret(command, "echo ${name}", args, parameters);
      });

      it("replaces the variable", () => {
        expect(result).toEqual("echo John");
      });
    });

    describe("with multiple variables and parameters", () => {
      beforeEach(() => {
        args = [];
        parameters = { firstName: "John", lastName: "Doe" };
        result = parameterInterpreter.interpret(command, "echo $firstName $lastName", args, parameters);
      });

      it("replaces the variable", () => {
        expect(result).toEqual("echo John Doe");
      });
    });
  });
});
