const ParameterInterpreter = require("../../lib/interpreters/ParameterInterpreter");
const parameterInterpreter = new ParameterInterpreter();

describe("parameterInterpreter", () => {
  describe("interpret", () => {
    let result, command, args, parameters;

    beforeEach(() => {
      command = {};
    });

    describe("without variables", () => {
      beforeEach(async () => {
        args = [];
        parameters = {};
        result = await parameterInterpreter.interpret(command, "mkdir test", args, parameters);
      });

      it("does not replace the text", () => {
        expect(result).toEqual("mkdir test");
      });
    });

    describe("with variable and no parameter", () => {
      beforeEach(async () => {
        args = [];
        parameters = {};
        result = await parameterInterpreter.interpret(command, "echo ${name}", args, parameters);
      });

      it("does not replace the variable", () => {
        expect(result).toEqual("echo ${name}");
      });
    });

    describe("with variable and parameter", () => {
      beforeEach(async () => {
        args = [];
        parameters = { name: "John" };
        result = await parameterInterpreter.interpret(command, "echo ${name}", args, parameters);
      });

      it("replaces the variable", () => {
        expect(result).toEqual("echo John");
      });
    });

    describe("with multiple variables and parameters", () => {
      beforeEach(async () => {
        args = [];
        parameters = { firstName: "John", lastName: "Doe" };
        result = await parameterInterpreter.interpret(command, "echo $firstName $lastName", args, parameters);
      });

      it("replaces the variable", () => {
        expect(result).toEqual("echo John Doe");
      });
    });
  });
});
