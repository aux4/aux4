const executorChain = require('../lib/executorChain');

const logExecutor = require('../lib/executors/logExecutor');
const profileExecutor = require('../lib/executors/profileExecutor');
const commandLineExecutor = require('../lib/executors/commandLineExecutor');

describe('executorChain', () => {
  describe('when there are only command line', () => {
    beforeEach(() => {
      logExecutor.execute = jest.fn().mockReturnValue(false);
      profileExecutor.execute = jest.fn().mockReturnValue(false);
      commandLineExecutor.execute = jest.fn().mockReturnValue(true);

    	executorChain.execute(['mkdir test', 'cd test']);
    });

    it('executes logExecutor for each command', () => {
      expect(logExecutor.execute.mock.calls.length).toEqual(2);
      expect(logExecutor.execute.mock.calls[0][0]).toEqual('mkdir test');
      expect(logExecutor.execute.mock.calls[1][0]).toEqual('cd test');
    });

    it('executes profileExecutor for each command', () => {
      expect(profileExecutor.execute.mock.calls.length).toEqual(2);
      expect(profileExecutor.execute.mock.calls[0][0]).toEqual('mkdir test');
      expect(profileExecutor.execute.mock.calls[1][0]).toEqual('cd test');
    });

    it('executes commandLineExecutor for each command', () => {
      expect(commandLineExecutor.execute.mock.calls.length).toEqual(2);
      expect(commandLineExecutor.execute.mock.calls[0][0]).toEqual('mkdir test');
      expect(commandLineExecutor.execute.mock.calls[1][0]).toEqual('cd test');
    });
  });

  describe('when there is a profile command', () => {
    beforeEach(() => {
      logExecutor.execute = jest.fn().mockReturnValue(false);
      profileExecutor.execute = jest.fn().mockReturnValue(true);
      commandLineExecutor.execute = jest.fn().mockReturnValue(true);

    	executorChain.execute(['profile:git']);
    });

    it('executes logExecutor for each command', () => {
      expect(logExecutor.execute.mock.calls.length).toEqual(1);
      expect(logExecutor.execute.mock.calls[0][0]).toEqual('profile:git');
    });

    it('executes profileExecutor for each command', () => {
      expect(profileExecutor.execute.mock.calls.length).toEqual(1);
      expect(profileExecutor.execute.mock.calls[0][0]).toEqual('profile:git');
    });

    it('executes commandLineExecutor for each command', () => {
      expect(commandLineExecutor.execute.mock.calls.length).toEqual(0);
    });
  });
});
