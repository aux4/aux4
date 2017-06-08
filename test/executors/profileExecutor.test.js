const profileExecutor = require('../../lib/executors/profileExecutor');

const executor = require('../../lib/executor');

describe('profileExecutor', () => {
  beforeEach(() => {
  	executor.defineProfile = jest.fn();
  	executor.execute = jest.fn();
  });

  describe('execute', () => {
    let command, args, result;

    describe('when is not a profile', () => {
      beforeEach(() => {
      	command = 'mkdir test';
        args = [];

        result = profileExecutor.execute(command, args);
      });

      it('returns false', () => {
        expect(result).toBeFalsy();
      });
    });

    describe('when is a profile', () => {
      let profile;

      beforeEach(() => {
        profile = 'git';
      	command = 'profile:' + profile;
        args = ['push'];

        result = profileExecutor.execute(command, args);
      });

      it('calls "executor.defineProfile" with the profile', () => {
        expect(executor.defineProfile).toHaveBeenCalledWith(profile);
      });

      it('executes "executor.execute"', () => {
        expect(executor.execute).toHaveBeenCalledWith(args);
      });

      it('returns true', () => {
        expect(result).toBeTruthy();
      });
    });
  });
});
