const logExecutor = require('../../lib/executors/logExecutor');

const out = require('../../lib/output');
const interpreter = require('../../lib/interpreter');

describe('logExecutor', () => {
  let spyOnInterpreter;

  beforeEach(() => {
  	out.println = jest.fn();
    spyOnInterpreter = jest.spyOn(interpreter, 'interpret');
  });

  describe('execute', () => {
    let action, args, parameters, result;

    describe('when action is not a log', () => {
      beforeEach(() => {
        action = 'mkdir test';
        args = [];
        parameters = {};

        result = logExecutor.execute({}, action, args, parameters);
      });

      it('returns false', () => {
        expect(result).toBeFalsy();
      });
    });

    describe('when action is a log', () => {
      describe('without parameters', () => {
        beforeEach(() => {
          action = 'log:mkdir test';
          args = [];
          parameters = {};

          result = logExecutor.execute({}, action, args, parameters);
        });

        it('prints the log', () => {
          expect(out.println).toHaveBeenCalledWith('mkdir test');
        });

        it('returns true', () => {
          expect(result).toBeTruthy();
        });
      });

      describe('with parameters', () => {
        beforeEach(() => {
          action = 'log:mkdir $folder';
          args = [];
          parameters = {folder: 'test'};

          result = logExecutor.execute({}, action, args, parameters);
        });

        it('calls the interpreter', () => {
          expect(spyOnInterpreter).toHaveBeenCalledWith('mkdir $folder', args, parameters);
        });

        it('prints the log', () => {
          expect(out.println).toHaveBeenCalledWith('mkdir test');
        });

        it('returns true', () => {
          expect(result).toBeTruthy();
        });
      });
    });
  });
});
