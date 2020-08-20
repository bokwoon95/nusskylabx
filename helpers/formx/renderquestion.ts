import m from "mithril";
import { Type, INode } from "./formx.d";
const SHA1 = require("crypto-js/sha1");

function idfy(...inputs: Array<string>): string {
  const sha1Bytes = SHA1(inputs.join(""));
  const output = sha1Bytes.toString();
  return output.substring(0, 8);
}

export function renderQuestion(node: INode): m.Vnode | Array<m.Vnode> {
  const renderfunc = rq.get(node.question.Type);
  if (!renderfunc) {
    return m("pre", `rendering function not found for question of type: ${node.question.Type}`);
  }
  return [m("h4", "Output"), renderfunc(node)];
}

const rq = new Map<Type, (node: INode) => m.Vnode>(); // render question
// rq Paragraph
rq.set(Type.Paragraph, function (node: INode): m.Vnode {
  return m("p", m("div", m.trust(node.question.Text)));
});
// rq Shorttext
rq.set(Type.Shorttext, function (node: INode): m.Vnode {
  return m(
    "p",
    m("div", m.trust(node.question.Text)),
    m("input.form-input.w-75", { type: "text", name: node.question.Name }),
  );
});
// rq Longtext
rq.set(Type.Longtext, function (node: INode): m.Vnode {
  return m(
    "p",
    m("div", m.trust(node.question.Text)),
    m(`textarea#${node.uuid}`, { name: node.question.Name, cols: 40, rows: 4 }),
  );
});
// rq Checkbox
rq.set(Type.Checkbox, function (node: INode): m.Vnode {
  return m(
    "p",
    m("div", m.trust(node.question.Text)),
    ...node.question.Options.map((option) =>
      m(
        "div",
        m("label.pointer", { for: idfy(node.question.Name, option.Value) }, [
          m("input.pointer.mr2", {
            type: "checkbox",
            name: node.question.Name,
            value: option.Value,
            id: idfy(node.question.Name, option.Value),
          }),
          option.Display,
        ]),
      ),
    ),
  );
});
// rq Select
rq.set(Type.Select, function (node: INode): m.Vnode {
  return m(
    "p",
    m("div", m.trust(node.question.Text)),
    m(
      "select.form-input",
      ...node.question.Options.map((option) => m("option", { value: option.Value }, option.Display)),
    ),
  );
});
// rq Radio
rq.set(Type.Radio, function (node: INode): m.Vnode {
  return m(
    "p",
    m("div", m.trust(node.question.Text)),
    ...node.question.Options.map((option) =>
      m(
        "div",
        m("label.pointer", { for: idfy(node.question.Name, option.Value) }, [
          m("input.pointer.mr2", {
            type: "radio",
            name: node.question.Name,
            value: option.Value,
            id: idfy(node.question.Name, option.Value),
          }),
          option.Display,
        ]),
      ),
    ),
  );
});
// rq Multiradio
rq.set(Type.Multiradio, function (node: INode): m.Vnode {
  return m(
    "p.overflow-x-scroll",
    m("div", m.trust(node.question.Text)),
    m(
      "table.multiradio",
      m(
        "tbody",
        m("tr", m("td"), ...node.question.Options.map((option) => m("td", option.Display))),
        ...node.question.Subquestions.map((subqn) =>
          m(
            "tr",
            m("td", subqn.Text),
            ...node.question.Options.map((option) =>
              m(
                "td.pv0",
                m(
                  "label.w-100.h-100.db.pointer",
                  { for: idfy(subqn.Name, option.Value) },
                  m("input.pointer", {
                    type: "radio",
                    name: subqn.Name,
                    value: option.Value,
                    id: idfy(subqn.Name, option.Value),
                  }),
                ),
              ),
            ),
          ),
        ),
      ),
    ),
  );
});
// rq Date
rq.set(Type.Date, function (node: INode): m.Vnode {
  return m("p", m("div", m.trust(node.question.Text)), m("input", { type: "date", name: node.question.Name }));
});
// rq Time
rq.set(Type.Time, function (node: INode): m.Vnode {
  return m("p", m("div", m.trust(node.question.Text)), m("input", { type: "time", name: node.question.Name }));
});
// rq Image
rq.set(Type.Image, function (node: INode): m.Vnode {
  return m(
    "p",
    m("div", m.trust(node.question.Text)),
    m("input", { type: "file", name: node.question.Name, accept: "image/*" }),
  );
});
