import m from "mithril";
import { Type, IQuestion, INode, EventHandler } from "./formx.d";

export function renderInput(node: INode, nodes: Array<INode>): m.Vnode | Array<m.Vnode> {
  const renderfunc = ri.get(node.question.Type);
  if (!renderfunc) {
    return m("pre", `rendering function not found for question of type: ${node.question.Type}`);
  }
  return [
    m("h4", "Input"),
    m(
      "div",
      m("div", "Type:"),
      m(
        "select",
        { onchange: changeQuestion(node, nodes) },
        Object.values(Type).map((qntype) =>
          m("option", { value: qntype, selected: qntype === node.question.Type }, qntype),
        ),
      ),
    ),
    m("div", renderfunc(node)),
  ];
}

const ri = new Map<Type, (node: INode) => m.Vnode>(); // render input
// ri Paragraph
ri.set(Type.Paragraph, function (node: INode): m.Vnode {
  return m(
    "div",
    m("p.pr4", m("div", "Text:"), m("textarea.w-100", { oninput: updateText(node), rows: 10 }, node.question.Text)),
  );
});
// ri Shorttext
ri.set(Type.Shorttext, function (node: INode): m.Vnode {
  return m(
    "div",
    m(
      "p.pr4",
      m("div", "Text:"),
      m("textarea.w-100", { oninput: updateText(node), cols: 40, rows: 10 }, node.question.Text),
    ),
    m(
      "p",
      m("div", m("span", "Name:")),
      m("input.db.w-80", {
        type: "text",
        oninput: updateName(node),
        value: node.question.Name,
        required: true,
        autocomplete: "off",
      }),
    ),
  );
});
// ri Longtext
ri.set(Type.Longtext, function (node: INode): m.Vnode {
  return m(
    "div",
    m(
      "p.pr4",
      m("div", "Text:"),
      m("textarea.w-100", { oninput: updateText(node), cols: 40, rows: 10 }, node.question.Text),
    ),
    m(
      "p",
      m("div", m("span", "Name:")),
      m("input.db.w-80", {
        type: "text",
        oninput: updateName(node),
        value: node.question.Name,
        required: true,
        autocomplete: "off",
      }),
    ),
  );
});
// ri Checkbox
ri.set(Type.Checkbox, function (node: INode): m.Vnode {
  return m(
    "div",
    m(
      "p.pr4",
      m("div", "Text:"),
      m("textarea.w-100", { oninput: updateText(node), cols: 40, rows: 10 }, node.question.Text),
    ),
    m(
      "p",
      m("div", m("span", "Name:")),
      m("input.db.w-80", {
        type: "text",
        oninput: updateName(node),
        value: node.question.Name,
        required: true,
        autocomplete: "off",
      }),
    ),
    m(
      "p",
      m("div", "Options:"),
      m("div", [
        m("button", { type: "button", onclick: addOption(node, -1) }, "add"),
        node.question.Options.map((option, i) =>
          m(
            "div",
            { key: i },
            m("div", `Option ${i+1} value:`),
            m("input", { type: "text", oninput: updateOptionValue(node, i), value: option.Value }),
            m("button", { type: "button", onclick: deleteOption(node, i) }, "delete"),
            m("button", { type: "button", onclick: addOption(node, i) }, "add"),
            m(
              "div",
              m("div", `Option ${i+1} label:`),
              m("textarea.w-90", { oninput: updateOptionDisplay(node, i), cols: 40, rows: 4, value: option.Display }),
            ),
          ),
        ),
      ]),
    ),
  );
});
// ri Select
ri.set(Type.Select, function (node: INode): m.Vnode {
  return m(
    "div",
    m(
      "p.pr4",
      m("div", "Text:"),
      m("textarea.w-100", { oninput: updateText(node), cols: 40, rows: 10 }, node.question.Text),
    ),
    m(
      "p",
      m("div", m("span", "Name:")),
      m("input.db.w-80", {
        type: "text",
        oninput: updateName(node),
        value: node.question.Name,
        required: true,
        autocomplete: "off",
      }),
    ),
    m(
      "p",
      m("div", "Options:"),
      m("div", [
        m("button", { type: "button", onclick: addOption(node, -1) }, "add"),
        node.question.Options.map((option, i) =>
          m(
            "div",
            { key: i },
            m("div", `Option ${i+1} value:`),
            m("input", { type: "text", oninput: updateOptionValue(node, i), value: option.Value }),
            m("button", { type: "button", onclick: deleteOption(node, i) }, "delete"),
            m("button", { type: "button", onclick: addOption(node, i) }, "add"),
            m(
              "div",
              m("div", `Option ${i+1} label:`),
              m("textarea.w-90", { oninput: updateOptionDisplay(node, i), cols: 40, rows: 4, value: option.Display }),
            ),
          ),
        ),
      ]),
    ),
  );
});
// ri Radio
ri.set(Type.Radio, function (node: INode): m.Vnode {
  return m(
    "div",
    m(
      "p.pr4",
      m("div", "Text:"),
      m("textarea.w-100", { oninput: updateText(node), cols: 40, rows: 10 }, node.question.Text),
    ),
    m(
      "p",
      m("div", m("span", "Name:")),
      m("input.db.w-80", {
        type: "text",
        oninput: updateName(node),
        value: node.question.Name,
        required: true,
        autocomplete: "off",
      }),
    ),
    m(
      "p",
      m("div", "Options:"),
      m("div", [
        m("button", { type: "button", onclick: addOption(node, -1) }, "add"),
        node.question.Options.map((option, i) =>
          m(
            "div",
            { key: i },
            m("div", `Option ${i+1} value:`),
            m("input", { type: "text", oninput: updateOptionValue(node, i), value: option.Value }),
            m("button", { type: "button", onclick: deleteOption(node, i) }, "delete"),
            m("button", { type: "button", onclick: addOption(node, i) }, "add"),
            m(
              "div",
              m("div", `Option ${i+1} label:`),
              m("textarea.w-90", { oninput: updateOptionDisplay(node, i), cols: 40, rows: 4, value: option.Display }),
            ),
          ),
        ),
      ]),
    ),
  );
});
// ri Multiradio
ri.set(Type.Multiradio, function (node: INode): m.Vnode {
  return m(
    "div",
    m(
      "p.pr4",
      m("div", "Text:"),
      m("textarea.w-100", { oninput: updateText(node), cols: 40, rows: 10 }, node.question.Text),
    ),
    m(
      "p",
      m("div", "Options:"),
      m("div", [
        m("button", { type: "button", onclick: addOption(node, -1) }, "add"),
        node.question.Options.map((option, i) =>
          m(
            "div",
            { key: i },
            m("div", `Option ${i+1} value:`),
            m("input", { type: "text", oninput: updateOptionValue(node, i), value: option.Value }),
            m("button", { type: "button", onclick: deleteOption(node, i) }, "delete"),
            m("button", { type: "button", onclick: addOption(node, i) }, "add"),
            m(
              "div",
              m("div", `Option ${i+1} label:`),
              m("textarea.w-90", { oninput: updateOptionDisplay(node, i), cols: 40, rows: 4, value: option.Display }),
            ),
          ),
        ),
      ]),
    ),
    m(
      "p",
      m(
        "p",
        m("div", m("span", "Subquestions:"), m("button", { type: "button", onclick: addSubquestion(node, -1) }, "add")),
        node.question.Subquestions.map((subqn, i) =>
          m(
            "div",
            { key: i },
            m(
              "p",
              m("div", m("span", "Name:")),
              m("input.db.w-80", {
                type: "text",
                oninput: updateSubquestionName(node, i),
                value: subqn.Name,
                required: true,
                autocomplete: "off",
              }),
              m("button", { type: "button", onclick: deleteSubquestion(node, i) }, "delete"),
              m("button", { type: "button", onclick: addSubquestion(node, i) }, "add"),
            ),
            m(
              "p.pr4",
              m("div", "Text:"),
              m("textarea.w-100", { oninput: updateSubquestion(node, i), cols: 40, rows: 10 }, subqn.Text),
            ),
          ),
        ),
      ),
    ),
  );
});
// ri Date
ri.set(Type.Date, function (node: INode): m.Vnode {
  return m(
    "div",
    m(
      "p.pr4",
      m("div", "Text:"),
      m("textarea.w-100", { oninput: updateText(node), cols: 40, rows: 10 }, node.question.Text),
    ),
    m(
      "p",
      m("div", m("span", "Name:")),
      m("input.db.w-80", {
        type: "text",
        oninput: updateName(node),
        value: node.question.Name,
        required: true,
        autocomplete: "off",
      }),
    ),
  );
});
// ri Time
ri.set(Type.Time, function (node: INode): m.Vnode {
  return m(
    "div",
    m(
      "p.pr4",
      m("div", "Text:"),
      m("textarea.w-100", { oninput: updateText(node), cols: 40, rows: 10 }, node.question.Text),
    ),
    m(
      "p",
      m("div", m("span", "Name:")),
      m("input.db.w-80", {
        type: "text",
        oninput: updateName(node),
        value: node.question.Name,
        required: true,
        autocomplete: "off",
      }),
    ),
  );
});
// ri Image
ri.set(Type.Image, function (node: INode): m.Vnode {
  return m(
    "div",
    m(
      "p.pr4",
      m("div", "Text:"),
      m("textarea.w-100", { oninput: updateText(node), cols: 40, rows: 10 }, node.question.Text),
    ),
    m(
      "p",
      m("div", m("span", "Name:")),
      m("input.db.w-80", {
        type: "text",
        oninput: updateName(node),
        value: node.question.Name,
        required: true,
        autocomplete: "off",
      }),
    ),
  );
});

function changeQuestion(node: INode, nodes: Array<INode>): EventHandler {
  return function (event: Event) {
    const el = event.currentTarget as HTMLSelectElement;
    if (!Object.values(Type).includes(el.value as Type)) {
      throw new Error(`${el.value} is not a valid type`);
    }
    const newquestion: IQuestion = {
      Type: el.value as Type,
      Text: node.question.Text,
      Name: node.question.Name,
      Options: node.question.Options || [],
      Subquestions: node.question.Subquestions || [],
    };
    const index = nodes.findIndex((x) => x.uuid === node.uuid);
    nodes[index].question = newquestion;
  };
}

//================//
// Event Handlers //
//================//

function updateName(node: INode): EventHandler {
  return function (event: Event) {
    const el = event.currentTarget as HTMLInputElement;
    node.question.Name = el.value;
  };
}

function updateText(node: INode): EventHandler {
  return function (event: Event) {
    const el = event.currentTarget as HTMLInputElement;
    node.question.Text = el.value;
  };
}

function updateOptionValue(node: INode, index: number): EventHandler {
  return function (event: Event) {
    const el = event.currentTarget as HTMLInputElement;
    node.question.Options[index].Value = el.value;
  };
}

function updateOptionDisplay(node: INode, index: number): EventHandler {
  return function (event: Event) {
    const el = event.currentTarget as HTMLInputElement;
    node.question.Options[index].Display = el.value;
  };
}

function addOption(node: INode, index: number): EventHandler {
  return function (_: Event) {
    // add empty string to index + 1
    node.question.Options.splice(index + 1, 0, { Value: "", Display: "" });
  };
}

function deleteOption(node: INode, index: number): EventHandler {
  return function (_: Event) {
    // delete value at index
    node.question.Options.splice(index, 1);
  };
}

function updateSubquestionName(node: INode, index: number): EventHandler {
  return function (event: Event) {
    const el = event.currentTarget as HTMLInputElement;
    node.question.Subquestions[index].Name = el.value;
  };
}

function updateSubquestion(node: INode, index: number): EventHandler {
  return function (event: Event) {
    const el = event.currentTarget as HTMLInputElement;
    node.question.Subquestions[index].Text = el.value;
  };
}

function addSubquestion(node: INode, index: number): EventHandler {
  return function (_: Event) {
    // add new subquestion to index + 1
    node.question.Subquestions.splice(index + 1, 0, { Name: "", Text: "" });
  };
}

function deleteSubquestion(node: INode, index: number): EventHandler {
  return function (_: Event) {
    // delete subquestion at index
    node.question.Subquestions.splice(index, 1);
  };
}
