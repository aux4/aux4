const executorChain = require('../lib/executorChain');

const logExecutor = require('../lib/executors/logExecutor');
const profileExecutor = require('../lib/executors/profileExecutor');
const commandLineExecutor = require('../lib/executors/commandLineExecutor');

describe('executorChain', () => {
  beforeEach(() => {
  	executorChain.add(logExecutor);
  	executorChain.add(profileExecutor);
  	executorChain.add(commandLineExecutor);
  });

  describe('when there are only command line', () => {
    let command, args, parameters;

    beforeEach(() => {
      logExecutor.execute = jest.fn().mockReturnValue(false);
      profileExecutor.execute = jest.fn().mockReturnValue(false);
      commandLineExecutor.execute = jest.fn().mockReturnValue(true);

      command = {
        execute: ['mkdir test', 'cd test']
      };
      args = [];
      parameters = {};
    	executorChain.execute(command, args, parameters);
    });

    it('executes logExecutor for each command', () => {
      expect(logExecutor.execute.mock.calls.length).toEqual(2);
      expect(logExecutor.execute.mock.calls[0][0]).toEqual(command);
      expect(logExecutor.execute.mock.calls[0][1]).toEqual('mkdir test');
      expect(logExecutor.execute.mock.calls[0][2]).toBe(args);
      expect(logExecutor.execute.mock.calls[0][3]).toBe(parameters);
      expect(logExecutor.execute.mock.calls[1][0]).toEqual(command);
      expect(logExecutor.execute.mock.calls[1][1]).toEqual('cd test');
      expect(logExecutor.execute.mock.calls[1][2]).toBe(args);
      expect(logExecutor.execute.mock.calls[1][3]).toBe(parameters);
    });

    it('executes profileExecutor for each command', () => {
      expect(profileExecutor.execute.mock.calls.length).toEqual(2);
      expect(profileExecutor.execute.mock.calls[0][0]).toEqual(command);
      expect(profileExecutor.execute.mock.calls[0][1]).toEqual('mkdir test');
      expect(profileExecutor.execute.mock.calls[0][2]).toBe(args);
      expect(profileExecutor.execute.mock.calls[0][3]).toBe(parameters);
      expect(profileExecutor.execute.mock.calls[1][0]).toEqual(command);
      expect(profileExecutor.execute.mock.calls[1][1]).toEqual('cd test');
      expect(profileExecutor.execute.mock.calls[1][2]).toBe(args);
      expect(profileExecutor.execute.mock.calls[1][3]).toBe(parameters);
    });

    it('executes commandLineExecutor for each command', () => {
      expect(commandLineExecutor.execute.mock.calls.length).toEqual(2);
      expect(commandLineExecutor.execute.mock.calls[0][0]).toEqual(command);
      expect(commandLineExecutor.execute.mock.calls[0][1]).toEqual('mkdir test');
      expect(commandLineExecutor.execute.mock.calls[0][2]).toBe(args);
      expect(commandLineExecutor.execute.mock.calls[0][3]).toBe(parameters);
      expect(commandLineExecutor.execute.mock.calls[1][0]).toEqual(command);
      expect(commandLineExecutor.execute.mock.calls[1][1]).toEqual('cd test');
      expect(commandLineExecutor.execute.mock.calls[1][2]).toBe(args);
      expect(commandLineExecutor.execute.mock.calls[1][3]).toBe(parameters);
    });
  });

  describe('when there is a profile command', () => {
    let command;

    beforeEach(() => {
      logExecutor.execute = jest.fn().mockReturnValue(false);
      profileExecutor.execute = jest.fn().mockReturnValue(true);
      commandLineExecutor.execute = jest.fn().mockReturnValue(true);

      command = {
        execute: ['profile:git']
      };
    	executorChain.execute(command);
    });

    it('executes logExecutor for each command', () => {
      expect(logExecutor.execute.mock.calls.length).toEqual(1);
      expect(logExecutor.execute.mock.calls[0][0]).toEqual(command);
      expect(logExecutor.execute.mock.calls[0][1]).toEqual('profile:git');
    });

    it('executes profileExecutor for each command', () => {
      expect(profileExecutor.execute.mock.calls.length).toEqual(1);
      expect(profileExecutor.execute.mock.calls[0][0]).toEqual(command);
      expect(profileExecutor.execute.mock.calls[0][1]).toEqual('profile:git');
    });

    it('executes commandLineExecutor for each command', () => {
      expect(commandLineExecutor.execute.mock.calls.length).toEqual(0);
    });
  });

  describe('when executor throws error', () => {
    beforeEach(() => {
    	logExecutor.execute = jest.fn().mockImplementation(() => {
        throw new Error();
      });
      profileExecutor.execute = jest.fn().mockReturnValue(true);
      commandLineExecutor.execute = jest.fn().mockReturnValue(true);

      command = {
        execute: ['profile:git']
      };

      executorChain.execute(command);
    });

    it('calls logExecutor', () => {
      expect(logExecutor.execute).toHaveBeenCalled();
    });

    it('does not call profileExecutor', () => {
      expect(profileExecutor.execute).not.toHaveBeenCalled();
    });

    it('does not call commandLineExecutor', () => {
      expect(commandLineExecutor.execute).not.toHaveBeenCalled();
    });
  });
});
