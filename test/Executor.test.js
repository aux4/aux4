const Config = require("../lib/Config");
const Profile = require("../lib/Profile");
const ExecutorChain = require("../lib/ExecutorChain");
const Help = require("../lib/Help");
const out = require("../lib/Output");
const Suggester = require("../lib/Suggester");

const Executor = require("../lib/Executor");

const config = new Config();
const executorChain = new ExecutorChain();

describe("executor", () => {
  let executor, configProfiles;
  beforeEach(() => {
    out.println = jest.fn();

    configProfiles = {
      profiles: [
        {
          name: "firstProfile",
          commands: [
            {
              value: "cmd",
              execute: ["mkdir first", "cd first"]
            }
          ]
        },
        {
          name: "secondProfile",
          commands: [
            {
              value: "cmd",
              execute: ["mkdir second", "cd second"]
            },
            {
              value: "cmd2",
              execute: ["mkdir second2", "cd second2"]
            },
            {
              value: "t",
              execute: ["mkdir t", "cd t"]
            }
          ]
        }
      ]
    };

    config.get = jest.fn().mockReturnValue(configProfiles);
    executor = new Executor(config, executorChain, Suggester);
  });

  describe("initialize executor", () => {
    it("calls config get", () => {
      expect(config.get).toHaveBeenCalled();
    });

    it("creates firstProfile", () => {
      expect(executor.profile("firstProfile").name()).toEqual("firstProfile");
    });

    it("creates secondProfile", () => {
      expect(executor.profile("secondProfile").name()).toEqual("secondProfile");
    });

    describe("current profile", () => {
      it('returns "main"', () => {
        expect(executor.currentProfile()).toEqual("main");
      });
    });
  });

  describe("change current profile", () => {
    describe("when profile does not exists", () => {
      it("throw an error", () => {
        expect(() => {
          executor.defineProfile("abc");
        }).toThrow(new Error("profile abc does not exists"));
      });
    });

    describe("when profile exists", () => {
      beforeEach(() => {
        executor.defineProfile("firstProfile");
      });

      describe("current profile", () => {
        it('returns "firstProfile"', () => {
          expect(executor.currentProfile()).toEqual("firstProfile");
        });
      });
    });
  });

  describe("execute", () => {
    describe("when there are no arguments", () => {
      beforeEach(() => {
        Help.print = jest.fn();
        executorChain.execute = jest.fn();

        executor = new Executor(config, executorChain, Suggester);
        executor.defineProfile("secondProfile");
        executor.execute([]);
      });

      it("prints help for each command", () => {
        expect(Help.print).toHaveBeenCalledWith(configProfiles.profiles[1].commands[0], 6);
      });
    });

    describe("when there are arguments", () => {
      let parameters;

      beforeEach(() => {
        executorChain.execute = jest.fn();

        parameters = { enable: "true" };

        executor = new Executor(config, executorChain, Suggester);
        executor.defineProfile("firstProfile");
        executor.execute(["cmd"], parameters);
      });

      it("calls executorChain", () => {
        expect(executorChain.execute).toHaveBeenCalledWith(configProfiles.profiles[0].commands[0], [], parameters);
      });
    });

    describe("when there are wrong arguments", () => {
      describe("with suggestion", () => {
        let suggester;

        beforeEach(() => {
          suggester = {};
          suggester.suggest = jest.fn();

          executor = new Executor(config, executorChain, suggester);
          executor.defineProfile("firstProfile");
          executor.execute(["c"], {});
        });

        it("calls suggest", () => {
          expect(suggester.suggest).toBeCalledWith(executor.profile("firstProfile"), "c");
        });
      });
    });

    describe("help", () => {
      beforeEach(() => {
        Help.print = jest.fn();

        executor = new Executor(config, executorChain, Suggester);
        executor.defineProfile("firstProfile");
        executor.execute(["cmd"], { help: true });
      });

      it("prints the help", () => {
        expect(Help.print).toHaveBeenCalledWith(configProfiles.profiles[0].commands[0], 5);
      });
    });
  });
});
