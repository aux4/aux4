const cryptr = require('cryptr');

describe('crypto', () => {
  let crypto, text, encryptedText;

  describe('without env variable', () => {
    beforeEach(() => {
      jest.resetModules();
    	crypto = require('../lib/crypto');

      text = 'this is the pure text';
      encryptedText = '56b7b26256902d91538ce19fdc8075f5a581789a26d1f5f07f794ef7964f5c63';
    });

    describe('encrypt', () => {
      it('returns the text encrypted', () => {
      	expect(crypto.encrypt(text)).toEqual(encryptedText);
      });
    });

    describe('decrypt', () => {
      it('returns the original text', () => {
        expect(crypto.decrypt(encryptedText)).toEqual(text);
      });
    });
  });

  describe('with env variable', () => {
    beforeEach(() => {
      jest.resetModules();
      process.env.AUX4_SECURITY_KEY = '1234';
    	crypto = require('../lib/crypto');

      text = 'this is the pure text';
      encryptedText = '503075d6009b8f0802950ed740d8d4a502546cabb4c77fbbee35997f378c2a52';
    });

    describe('encrypt', () => {
      it('returns the text encrypted', () => {
      	expect(crypto.encrypt(text)).toEqual(encryptedText);
      });
    });

    describe('decrypt', () => {
      it('returns the original text', () => {
        expect(crypto.decrypt(encryptedText)).toEqual(text);
      });
    });
  });
});
