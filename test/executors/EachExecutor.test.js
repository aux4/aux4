const EachExecutor = require("../../lib/executors/EachExecutor");

describe("EachExecutor", () => {
  let eachExecutor, executorChain, interpreter;

  beforeEach(() => {
    interpreter = {};
    interpreter.interpret = jest.fn().mockReturnValue("command");

    executorChain = {};
    executorChain.execute = jest.fn();

    eachExecutor = new EachExecutor(interpreter, executorChain);
  });

  describe("execute", () => {
    let action, args, parameters, result;

    beforeEach(() => {
      action = "mkdir folder";
      args = [];
    });

    describe("when action does not start with each:", () => {
      beforeEach(async () => {
        action = "mkdir test";
        result = await eachExecutor.execute("command", action, args, parameters);
      });

      it("returns false", () => {
        expect(result).toEqual(false);
      });
    });

    describe("when action starts with each:", () => {
      describe("when response not iterable", () => {
        describe("when response is an array", () => {
          beforeEach(async () => {
            action = "each:mkdir $item";
            parameters = { response: ["folder1", "folder2"] };
            result = await eachExecutor.execute("command", action, args, parameters);
          });

          it("returns true", () => {
            expect(result).toEqual(true);
          });

          it("calls interpreter interpret", () => {
            expect(interpreter.interpret).toHaveBeenCalledTimes(2);
            expect(interpreter.interpret).toHaveBeenCalledWith("command", "mkdir $item", [], {
              item: "folder2",
              ...parameters
            });
            expect(interpreter.interpret).toHaveBeenCalledWith("command", "mkdir $item", [], {
              item: "folder2",
              ...parameters
            });
          });

          it("calls executorChain execute", () => {
            expect(executorChain.execute).toHaveBeenCalledTimes(2);
            expect(executorChain.execute).toHaveBeenCalledWith({ execute: ["command"] }, [], {
              item: "folder1",
              ...parameters
            });
            expect(executorChain.execute).toHaveBeenCalledWith({ execute: ["command"] }, [], {
              item: "folder2",
              ...parameters
            });
          });
        });

        describe("when response is a string", () => {
          beforeEach(async () => {
            action = "each:mkdir $item";
            parameters = { response: "folder1\nfolder2" };
            result = await eachExecutor.execute("command", action, args, parameters);
          });

          it("returns true", () => {
            expect(result).toEqual(true);
          });

          it("calls interpreter interpret", () => {
            expect(interpreter.interpret).toHaveBeenCalledTimes(2);
            expect(interpreter.interpret).toHaveBeenCalledWith("command", "mkdir $item", [], {
              item: "folder2",
              ...parameters
            });
            expect(interpreter.interpret).toHaveBeenCalledWith("command", "mkdir $item", [], {
              item: "folder2",
              ...parameters
            });
          });

          it("calls executorChain execute", () => {
            expect(executorChain.execute).toHaveBeenCalledTimes(2);
            expect(executorChain.execute).toHaveBeenCalledWith({ execute: ["command"] }, [], {
              item: "folder1",
              ...parameters
            });
            expect(executorChain.execute).toHaveBeenCalledWith({ execute: ["command"] }, [], {
              item: "folder2",
              ...parameters
            });
          });
        });

        describe("when response is not an array or string", () => {
          beforeEach(() => {
            action = "each:mkdir $item";
            parameters = { response: {} };
          });

          it("throws error", async () => {
            await expect(() => eachExecutor.execute("command", action, args, parameters)).rejects.toThrow(
              "response is not iterable"
            );
          });
        });
      });
    });
  });
});
