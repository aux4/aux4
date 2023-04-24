const Variable = require("../lib/Variable");

describe("variable", () => {
  describe("list", () => {
    describe("when action is not a string", () => {
      let list, action;

      beforeEach(() => {
        action = 123;

        list = Variable.list(action);
      });

      it("returns an empty list", () => {
        expect(list).toEqual([]);
      });
    });

    describe("without variables", () => {
      let list, action;

      beforeEach(() => {
        action = "mkdir folder";

        list = Variable.list(action);
      });

      it("returns an empty list", () => {
        expect(list).toEqual([]);
      });
    });

    describe("single variable", () => {
      let list, action;

      beforeEach(() => {
        action = "mkdir ${folder}";

        list = Variable.list(action);
      });

      it("returns folder in the list", () => {
        expect(list).toEqual(["folder"]);
      });
    });

    describe("duplicated variable", () => {
      let list, action;

      beforeEach(() => {
        action = "echo ${name} / $name";

        list = Variable.list(action);
      });

      it("returns name in the list", () => {
        expect(list).toEqual(["name"]);
      });
    });

    describe("multiple variables", () => {
      let list, action;

      beforeEach(() => {
        action = "echo ${firstName} ${middleName} $lastName";

        list = Variable.list(action);
      });

      it("returns first, middle and last name in the list", () => {
        expect(list).toEqual(["firstName", "middleName", "lastName"]);
      });
    });
  });

  describe("replace", () => {
    describe("single variable", () => {
      let result, action, key, value;

      beforeEach(() => {
        action = "mkdir ${folder}";
        key = "folder";
        value = "test";

        result = Variable.replace(action, key, value);
      });

      it("replaces variable", () => {
        expect(result).toEqual("mkdir test");
      });
    });

    describe("duplicated variable", () => {
      let result, action, key, value;

      beforeEach(() => {
        action = "echo ${name} / $name";
        key = "name";
        value = "test";

        result = Variable.replace(action, key, value);
      });

      it("replaces variables", () => {
        expect(result).toEqual("echo test / test");
      });
    });

    describe("multiple variables", () => {
      let result, action, key, value;

      beforeEach(() => {
        action = "echo ${firstName} ${lastName}";
        key = "firstName";
        value = "John";

        result = Variable.replace(action, key, value);
      });

      it("replaces $firstName", () => {
        expect(result).toEqual("echo John ${lastName}");
      });
    });

    describe("value as object", () => {
      let result, action, key, value;

      beforeEach(() => {
        value = { firstName: "John", lastName: "Doe" };
      });

      describe("when key is defined", () => {
        beforeEach(() => {
          action = "echo ${person.firstName} $person.lastName";
          key = "person";

          result = Variable.replace(action, key, value);
        });

        it("replaces ${person.firstName} and $person.lastName", () => {
          expect(result).toEqual("echo John Doe");
        });
      });

      describe("when key is not defined", () => {
        beforeEach(() => {
          action = "echo ${person.address.city}";
          key = "person";

          result = Variable.replace(action, key, value);
        });

        it("replaces ${person.address.city}", () => {
          expect(result).toEqual("echo ");
        });
      });
    });

    describe("action is just a variable", () => {
      let result, action, key, value;

      beforeEach(() => {
        action = "${folder}";
        key = "folder";
        value = "test";

        result = Variable.replace(action, key, value);
      });

      it("replaces variable", () => {
        expect(result).toEqual("test");
      });
    });

    describe("variable with index", () => {
      let result, action, key, value;

      describe("when array is in the object", () => {
        beforeEach(() => {
          action = "echo ${main.folder[1]}";
          key = "main";
          value = { folder: ["test1", "test2"] };

          result = Variable.replace(action, key, value);
        });

        it("replaces variable", () => {
          expect(result).toEqual("echo test2");
        });
      });

      describe("when array is not in the object", () => {
        beforeEach(() => {
          action = "echo ${folder[1]}";
          key = "folder";
          value = ["test1", "test2"];

          result = Variable.replace(action, key, value);
        });

        it("replaces variable", () => {
          expect(result).toEqual("echo test2");
        });
      });
    });
  });
});
