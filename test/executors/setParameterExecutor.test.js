const setParameterExecutor = require('../../lib/executors/setParameterExecutor');

const out = require('../../lib/output');
const interpreter = require('../../lib/interpreter');
const parameterInterpreter = require('../../lib/interpreters/parameterInterpreter');

describe('setParameterExecutor', () => {
  let spyOnInterpreter;

  beforeEach(() => {
    out.println = jest.fn();
    interpreter.add(parameterInterpreter);
    spyOnInterpreter = jest.spyOn(interpreter, 'interpret');
  });

  describe('execute', () => {
    let command, action, args, parameters, result;

    describe('when action is not a set', () => {
      beforeEach(() => {
        action = 'mkdir test';
        args = [];
        parameters = {};

        result = setParameterExecutor.execute({}, action, args, parameters);
      });

      it('returns false', () => {
        expect(result).toBeFalsy();
      });
    });

    describe('when action is a set', () => {
      describe('without equals', () => {
        beforeEach(() => {
          action = 'set:variable';
          args = [];
          parameters = {};

          result = setParameterExecutor.execute({}, action, args, parameters);
        });

        it('prints the log', () => {
          expect(out.println).toHaveBeenCalledWith('The set format is: set:<param-name>=<param-value>'.red);
        });

        it('returns true', () => {
          expect(result).toBeTruthy();
        });
      });

      describe('with equals ', () => {
        describe('and static value', () => {
          beforeEach(() => {
            command = {};
            action = 'set:variable=value';
            args = [];
            parameters = {};

            result = setParameterExecutor.execute(command, action, args, parameters);
          });

          it('calls the interpreter', () => {
            expect(spyOnInterpreter).toHaveBeenCalledWith(command, 'value', args, parameters);
          });

          it('sets parameter to the context', () => {
            expect(parameters.variable).toBe('value');
          });

          it('returns true', () => {
            expect(result).toBeTruthy();
          });
        });

        describe('and value from another parameter', () => {
          beforeEach(() => {
            command = {};
            action = 'set:name=${firstName} ${lastName}';
            args = [];
            parameters = {
              firstName: 'John',
              lastName: 'Doe'
            };

            result = setParameterExecutor.execute(command, action, args, parameters);
          });

          it('calls the interpreter', () => {
            expect(spyOnInterpreter).toHaveBeenCalledWith(command, '${firstName} ${lastName}', args, parameters);
          });

          it('sets parameter to the context', () => {
            expect(parameters.name).toBe('John Doe');
          });

          it('returns true', () => {
            expect(result).toBeTruthy();
          });
        });
      });
    });
  });
});
