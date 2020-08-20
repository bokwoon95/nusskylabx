import m from "mithril";
import { Type, FormStatus, INode, EventHandler } from "./formx.d";
import { renderInput } from "./renderinput";
import { renderQuestion } from "./renderquestion";
import { isIQuestionArray } from "./typecheck";
import { hideElement } from "../../app/utils";
const uuidv4 = require("uuid/v4");
let highlightedQuestion: string = "";
let formstatus: FormStatus = FormStatus.ViewForm;

function tryParse(input: any): object {
  try {
    return JSON.parse(input);
  } catch {
    return undefined;
  }
}

function nodesToString(nodes: Array<INode>): string {
  return JSON.stringify(nodes.map((x) => x.question))
    .replace(/</g, "\\u003c")
    .replace(/>/g, "\\u003e")
    .replace(/&/g, "\\u0026");
}

export function initFormbuilder({ selector, data }: { selector: string; data: string }) {
  const textarea = document.querySelector(selector);
  if (!(textarea instanceof HTMLTextAreaElement)) {
    throw new Error(`${selector} is not a textarea`);
  }
  hideElement(textarea);
  let questions = tryParse(data);
  if (questions === null || questions === undefined) {
    questions = [];
  }
  const nodes = new Array<INode>();
  if (isIQuestionArray(questions)) {
    for (let question of questions) {
      nodes.push({ uuid: uuidv4(), question: question });
    }
    console.log(nodes);
  } else {
    throw new Error(`${data} is not an IQuestion Array`);
  }
  const mountpoint = document.createElement("div");
  mountpoint.setAttribute(
    "id",
    Math.random().toString(36).substring(2, 15) + Math.random().toString(36).substring(2, 15),
  );
  textarea.parentNode.insertBefore(mountpoint, textarea);
  textarea.form.addEventListener("submit", function (_: Event) {
    (textarea as HTMLTextAreaElement).value = nodesToString(nodes);
  });
  m.mount(mountpoint, { view: () => Formbuilder(nodes, data) });
}

function Formbuilder(nodes: Array<INode>, data: string): Array<m.Vnode | Array<m.Vnode>> {
  if (nodes === null || nodes === undefined) {
    throw new Error("nodes is undefined");
  }
  const questions = nodes.map((x) => x.question);
  const viewJSONTextarea = m(
    "div.pt2",
    m(
      "textarea.w-100.b--black",
      {
        rows: 20,
        readonly: true,
      },
      JSON.stringify(questions),
    ),
  );
  const loadJSONTextarea: m.Vnode = m(
    "div.pt2",
    m("textarea.w-100.b--black", {
      id: "loadJSON",
      rows: 10,
    }),
    m("button", { type: "button", onclick: loadJSON(nodes) }, "Load"),
  );
  switch (formstatus) {
    case FormStatus.ViewJSON:
      return [Tabbar(nodes, data), viewJSONTextarea];
    case FormStatus.LoadJSON:
      return [Tabbar(nodes, data), loadJSONTextarea];
    case FormStatus.ViewForm:
    // fallthrough
    default:
      return [Tabbar(nodes, data), Questions(nodes)];
  }
}

function Tabbar(nodes: Array<INode>, data: string): Array<m.Vnode> {
  // Declare buttons/components
  const buttonAdd: m.Vnode = m("button.ph2", { type: "button", onclick: addQuestion(null, nodes) }, "add");
  const buttonDeleteAll: m.Vnode = m(
    "button.ph2",
    {
      type: "button",
      onclick: () => {
        nodes.splice(0, nodes.length);
        formstatus = FormStatus.ViewForm;
      },
    },
    "delete all",
  );
  const buttonViewHideJSON: m.Vnode = m(
    "button.ph2",
    {
      type: "button",
      onclick: () => {
        if (formstatus === FormStatus.ViewJSON) {
          formstatus = FormStatus.ViewForm;
        } else {
          formstatus = FormStatus.ViewJSON;
        }
      },
    },
    formstatus === FormStatus.ViewJSON ? "hide JSON" : "view JSON",
  );
  const spacer: m.Vnode = m("span.mh2");
  const isModified = nodesToString(nodes) !== data;
  const buttonReset: m.Vnode = m(
    "button.ph2",
    {
      type: "button",
      disabled: !isModified,
      onclick: () => {
        resetNodes(nodes, data);
      },
    },
    "Reset",
  );
  const buttonLoadJSON: m.Vnode = m(
    "button.ph2",
    {
      type: "button",
      onclick: () => {
        formstatus = FormStatus.LoadJSON;
      },
    },
    "load from JSON",
  );
  // Assemble the tabbar depending on whether there are any nodes present. Each
  // node represents one question.
  const tabbar: Array<m.Vnode> = [buttonAdd];
  if (nodes.length !== 0) {
    tabbar.push(buttonDeleteAll, buttonViewHideJSON, spacer, buttonReset);
  } else {
    tabbar.push(buttonLoadJSON, spacer, buttonReset);
  }
  return tabbar;
}

/**
 * Render the list of questions from a list of nodes.
 */
function Questions(nodes: Array<INode>): Array<m.Vnode> {
  return nodes.map((node, i) =>
    m(
      "div.mv2.pa2.ba" + (highlightedQuestion === node.uuid ? ".bg-light-gray" : ".bg-white"),
      { key: node.uuid },
      m(
        "div.flex.justify-between",
        m("div", `Q${i + 1}.`),
        m(
          "div",
          m("button.ph2", { type: "button", onclick: moveQuestionUp(node, nodes) }, "up"),
          m("button.ph2", { type: "button", onclick: moveQuestionDown(node, nodes) }, "down"),
        ),
      ),

      // This line is where the left hand side (form input) and right hand side
      // (question output) is rendered.
      m("div.flex-l", m("div.w-50-l", renderInput(node, nodes)), m("div.w-50-l", renderQuestion(node))),

      m(
        "div.flex.justify-end",
        m("button.ph2", { type: "button", onclick: addQuestion(node, nodes) }, "add"),
        m("button.ph2", { type: "button", onclick: deleteQuestion(node, nodes) }, "delete"),
      ),
    ),
  );
}

//================//
// Event Handlers //
//================//

/**
 * resetNodes is a helper function that deletes all existing question nodes and
 * loads them back in with the initial dataString (back to factory settings so
 * to speak). dataString should never change over the lifetime of the
 * application, barring a full page refresh.
 */
function resetNodes(nodes: Array<INode>, dataString: string): void {
  nodes.splice(0, nodes.length); // Delete existing nodes
  const questions = tryParse(dataString);
  if (questions === null || questions === undefined) {
    throw new Error(`could not parse json: ${dataString}`);
  }
  if (!isIQuestionArray(questions)) {
    throw new Error(`json is not an IQuestion array: ${dataString}`);
  }
  for (const question of questions) {
    nodes.push({ uuid: uuidv4(), question: question });
  }
}

/**
 * loadJSON(nodes) returns an event handler that grabs the text in a #loadJSON
 * textarea and tries to parse it as an Array<IQuestion>. If successful, it
 * will load push Array<IQuestion> contents into nodes
 */
function loadJSON(nodes: Array<INode>) {
  return function (_: Event) {
    const loadJSONTextarea: HTMLTextAreaElement = document.querySelector("#loadJSON");
    if (loadJSONTextarea === null || loadJSONTextarea === undefined) {
      console.log('document.querySelector("#loadJSON") returned undefined');
      return;
    }
    if (loadJSONTextarea instanceof HTMLTextAreaElement === false) {
      console.log(`${loadJSONTextarea} is not a textarea`);
      return;
    }
    resetNodes(nodes, loadJSONTextarea.value);
    formstatus = FormStatus.ViewForm;
  };
}

/**
 *
 */
function addQuestion(node: INode | null, nodes: Array<INode>): EventHandler {
  return function (_: Event) {
    if (formstatus === FormStatus.LoadJSON) {
      formstatus = FormStatus.ViewForm;
    }
    const index = node !== null && node !== undefined ? nodes.findIndex((x) => x.uuid === node.uuid) : -1;
    if (index === null || index === undefined) {
      console.log(`index not found: ${index}`);
      return;
    }
    const uuid = uuidv4();
    highlightedQuestion = uuid;
    // add new node at index + 1
    nodes.splice(index + 1, 0, {
      uuid: uuid,
      question: {
        Type: Type.Shorttext,
        Name: "",
        Text: "",
      },
    });
    console.log(nodes);
  };
}

function deleteQuestion(node: INode, nodes: Array<INode>): EventHandler {
  return function (_: Event) {
    const index = nodes.findIndex((x) => x.uuid === node.uuid);
    if (index === null || index === undefined) {
      console.log(`index not found: ${index}`);
      return;
    }
    highlightedQuestion = "";
    // delete node at index
    nodes.splice(index, 1);
  };
}

function moveQuestionUp(node: INode, nodes: Array<INode>): EventHandler {
  return function (_: Event) {
    const index = nodes.findIndex((x) => x.uuid === node.uuid);
    if (index === null || index === undefined) {
      console.log(`index not found: ${index}`);
      return;
    }
    highlightedQuestion = node.uuid;
    // swap the elements at index and index - 1 (but only if index exists and is within bounds)
    if (index > 0) {
      [nodes[index], nodes[index - 1]] = [nodes[index - 1], nodes[index]];
    }
  };
}

function moveQuestionDown(node: INode, nodes: Array<INode>): EventHandler {
  return function (_: Event) {
    const index = nodes.findIndex((x) => x.uuid === node.uuid);
    if (index === null || index === undefined) {
      console.log(`index not found: ${index}`);
      return;
    }
    highlightedQuestion = node.uuid;
    // swap the elements at index and index + 1 (but only if index exists and is within bounds)
    if (index < nodes.length - 1) {
      [nodes[index], nodes[index + 1]] = [nodes[index + 1], nodes[index]];
    }
  };
}
