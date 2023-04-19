const colors = require("colors");

const ExecutorChain = require("../lib/ExecutorChain");

const out = require("../lib/Output");

jest.mock("../lib/executors/LogExecutor");
const LogExecutor = require("../lib/executors/LogExecutor");

const ProfileExecutor = require("../lib/executors/ProfileExecutor");

jest.mock("../lib/executors/CommandLineExecutor");
const CommandLineExecutor = require("../lib/executors/CommandLineExecutor");

describe("executorChain", () => {
  let executorChain, logExecutor, profileExecutor, commandLineExecutor;

  beforeEach(() => {
    out.println = jest.fn();

    logExecutor = {
      execute: jest.fn()
    };
    LogExecutor.mockReturnValue(logExecutor);

    profileExecutor = {
      execute: jest.fn()
    };
    ProfileExecutor.with = jest.fn().mockReturnValue(jest.fn().mockReturnValue(profileExecutor));

    commandLineExecutor = {
      execute: jest.fn()
    };
    CommandLineExecutor.mockReturnValue(commandLineExecutor);

    executorChain = new ExecutorChain();
    executorChain.register(LogExecutor);
    executorChain.register(ProfileExecutor.with(null));
    executorChain.register(CommandLineExecutor);
  });

  describe("when execute is not defined", () => {
    let command, args, parameters;

    beforeEach(() => {
      command = {};
      args = [];
      parameters = {};

      executorChain.execute(command, args, parameters);
    });

    it('prints "execute is not defined"', () => {
      expect(out.println).toHaveBeenCalledWith("execute is not defined".red);
    });
  });

  describe("when execute is a function", () => {
    let command, args, parameters;

    beforeEach(async () => {
      command = {
        execute: jest.fn()
      };
      args = [];
      parameters = {};

      await executorChain.execute(command, args, parameters);
    });

    it("calls execute", () => {
      expect(command.execute).toHaveBeenCalledWith(parameters, args, command);
    });
  });

  describe("when there are only command line", () => {
    let command, args, parameters;

    beforeEach(async () => {
      logExecutor.execute = jest.fn().mockReturnValue(false);
      profileExecutor.execute = jest.fn().mockReturnValue(false);
      commandLineExecutor.execute = jest.fn().mockReturnValue(true);

      command = {
        execute: ["mkdir test", "cd test"]
      };
      args = [];
      parameters = {};
      await executorChain.execute(command, args, parameters);
    });

    it("executes logExecutor for each command", () => {
      expect(logExecutor.execute.mock.calls.length).toEqual(2);
      expect(logExecutor.execute.mock.calls[0][0]).toEqual(command);
      expect(logExecutor.execute.mock.calls[0][1]).toEqual("mkdir test");
      expect(logExecutor.execute.mock.calls[0][2]).toBe(args);
      expect(logExecutor.execute.mock.calls[0][3]).toBe(parameters);
      expect(logExecutor.execute.mock.calls[1][0]).toEqual(command);
      expect(logExecutor.execute.mock.calls[1][1]).toEqual("cd test");
      expect(logExecutor.execute.mock.calls[1][2]).toBe(args);
      expect(logExecutor.execute.mock.calls[1][3]).toBe(parameters);
    });

    it("executes profileExecutor for each command", () => {
      expect(profileExecutor.execute.mock.calls.length).toEqual(2);
      expect(profileExecutor.execute.mock.calls[0][0]).toEqual(command);
      expect(profileExecutor.execute.mock.calls[0][1]).toEqual("mkdir test");
      expect(profileExecutor.execute.mock.calls[0][2]).toBe(args);
      expect(profileExecutor.execute.mock.calls[0][3]).toBe(parameters);
      expect(profileExecutor.execute.mock.calls[1][0]).toEqual(command);
      expect(profileExecutor.execute.mock.calls[1][1]).toEqual("cd test");
      expect(profileExecutor.execute.mock.calls[1][2]).toBe(args);
      expect(profileExecutor.execute.mock.calls[1][3]).toBe(parameters);
    });

    it("executes commandLineExecutor for each command", () => {
      expect(commandLineExecutor.execute.mock.calls.length).toEqual(2);
      expect(commandLineExecutor.execute.mock.calls[0][0]).toEqual(command);
      expect(commandLineExecutor.execute.mock.calls[0][1]).toEqual("mkdir test");
      expect(commandLineExecutor.execute.mock.calls[0][2]).toBe(args);
      expect(commandLineExecutor.execute.mock.calls[0][3]).toBe(parameters);
      expect(commandLineExecutor.execute.mock.calls[1][0]).toEqual(command);
      expect(commandLineExecutor.execute.mock.calls[1][1]).toEqual("cd test");
      expect(commandLineExecutor.execute.mock.calls[1][2]).toBe(args);
      expect(commandLineExecutor.execute.mock.calls[1][3]).toBe(parameters);
    });
  });

  describe("when there is a profile command", () => {
    let command;

    beforeEach(async () => {
      logExecutor.execute = jest.fn().mockReturnValue(false);
      profileExecutor.execute = jest.fn().mockReturnValue(true);
      commandLineExecutor.execute = jest.fn().mockReturnValue(true);

      command = {
        execute: ["profile:git"]
      };
      await executorChain.execute(command);
    });

    it("executes logExecutor for each command", () => {
      expect(logExecutor.execute.mock.calls.length).toEqual(1);
      expect(logExecutor.execute.mock.calls[0][0]).toEqual(command);
      expect(logExecutor.execute.mock.calls[0][1]).toEqual("profile:git");
    });

    it("executes profileExecutor for each command", () => {
      expect(profileExecutor.execute.mock.calls.length).toEqual(1);
      expect(profileExecutor.execute.mock.calls[0][0]).toEqual(command);
      expect(profileExecutor.execute.mock.calls[0][1]).toEqual("profile:git");
    });

    it("executes commandLineExecutor for each command", () => {
      expect(commandLineExecutor.execute.mock.calls.length).toEqual(0);
    });
  });

  describe("when executor throws error", () => {
    beforeEach(async () => {
      logExecutor.execute = jest.fn().mockImplementation(() => {
        throw new Error();
      });
      profileExecutor.execute = jest.fn().mockReturnValue(true);
      commandLineExecutor.execute = jest.fn().mockReturnValue(true);

      command = {
        execute: ["profile:git"]
      };

      await executorChain.execute(command);
    });

    it("calls logExecutor", () => {
      expect(logExecutor.execute).toHaveBeenCalled();
    });

    it("does not call profileExecutor", () => {
      expect(profileExecutor.execute).not.toHaveBeenCalled();
    });

    it("does not call commandLineExecutor", () => {
      expect(commandLineExecutor.execute).not.toHaveBeenCalled();
    });
  });
});
