const LogExecutor = require("../../lib/executors/LogExecutor");

const out = require("../../lib/Output");
const Interpreter = require("../../lib/Interpreter");
const ParameterInterpreter = require("../../lib/interpreters/ParameterInterpreter");

const interpreter = new Interpreter();
const logExecutor = new LogExecutor(interpreter);

describe("logExecutor", () => {
  let spyOnInterpreter;

  beforeEach(() => {
    out.println = jest.fn();
    interpreter.add(new ParameterInterpreter());
    spyOnInterpreter = jest.spyOn(interpreter, "interpret");
  });

  describe("execute", () => {
    let command, action, args, parameters, result;

    describe("when action is not a log", () => {
      beforeEach(() => {
        action = "mkdir test";
        args = [];
        parameters = {};

        result = logExecutor.execute({}, action, args, parameters);
      });

      it("returns false", () => {
        expect(result).toBeFalsy();
      });
    });

    describe("when action is a log", () => {
      describe("without parameters", () => {
        beforeEach(() => {
          action = "log:mkdir test";
          args = [];
          parameters = {};

          result = logExecutor.execute({}, action, args, parameters);
        });

        it("prints the log", () => {
          expect(out.println).toHaveBeenCalledWith("mkdir test");
        });

        it("returns true", () => {
          expect(result).toBeTruthy();
        });
      });

      describe("with parameters", () => {
        beforeEach(() => {
          command = {};
          action = "log:mkdir $folder";
          args = [];
          parameters = { folder: "test" };

          result = logExecutor.execute(command, action, args, parameters);
        });

        it("calls the interpreter", () => {
          expect(spyOnInterpreter).toHaveBeenCalledWith(command, "mkdir $folder", args, parameters);
        });

        it("prints the log", () => {
          expect(out.println).toHaveBeenCalledWith("mkdir test");
        });

        it("returns true", () => {
          expect(result).toBeTruthy();
        });
      });
    });
  });
});
