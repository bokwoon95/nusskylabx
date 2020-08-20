const SHA1 = require("crypto-js/sha1");

export const log = process.env.IS_PROD === "true" ? (..._: any) => {} : console.log;

export function tryParse(input: any): object {
  try {
    return JSON.parse(input);
  } catch {
    return undefined;
  }
}

export function getCsrfToken(): string {
  // gorilla.csrf.Token is the default csrf input name by the gorilla csrf library
  const tokens = document.getElementsByName("gorilla.csrf.Token");
  if (tokens.length <= 0) {
    console.log(
      `Unable to find gorilla csrf DOM element on page. ` +
        `To fix this, please include the text '{{SkylabCsrfToken}}' anywhere within the <head> or the <body> tags of ` +
        `the HTML page on the server side.`,
    );
    return "";
  }
  const token = (tokens[0] as HTMLInputElement).value;
  if (!token) {
    throw new Error(`Found gorilla csrf DOM element ${token} but value not found inside`);
  }
  return token;
}

export function idfy(...inputs: Array<string>): string {
  const sha1Bytes = SHA1(inputs.join(""));
  const output = sha1Bytes.toString();
  return output.substring(0, 8);
}

export function hideElement(el: Element) {
  el.classList.remove("dib");
  el.classList.add("dn");
}
export function showElement(el: Element) {
  el.classList.remove("dn");
  el.classList.add("dib");
}

export function datatablesClickHandler(selected: Set<string>, redraw: () => void): (event: Event) => void {
  return function (event: Event) {
    const target = event.target as HTMLElement;
    if (target.tagName === "A") {
      return;
    }
    if (target.classList.contains("dataTables_empty")) {
      return;
    }
    $(this).toggleClass("selected");
    const fsid = $(this).find("td:first").text();
    if (selected.has(fsid)) {
      selected.delete(fsid);
      console.log(selected);
    } else {
      selected.add(fsid);
      console.log(selected);
    }
    if (redraw !== null && redraw != undefined) {
      redraw();
    }
  };
}

export function datatablesSelectAll(selected: Set<string>, redraw: () => void): (event: Event) => void {
  return function (_: Event) {
    const rows = $(`tbody > [role=row]`);
    rows.addClass("selected");
    selected.clear();
    rows.each((_, el) => {
      selected.add(el.querySelector("td").innerText);
    });
    console.log(selected);
    redraw();
  };
}

export function datatablesUnselectAll(selected: Set<string>, redraw: () => void): (event: Event) => void {
  return function (_: Event) {
    const rows = $(`tbody > [role=row]`);
    rows.removeClass("selected");
    selected.clear();
    console.log(selected);
    redraw();
  };
}

export function renderSelectedAsInputsIntoElement(selected: Set<string>, el: Element, name: string) {
  while (el.lastElementChild) {
    el.lastElementChild.remove();
  }
  for (const item of selected) {
    const input = document.createElement("input");
    input.type = "hidden";
    input.name = name;
    input.value = item.toString().trim();
    el.appendChild(input);
  }
}
