import {
  datatablesClickHandler,
  datatablesSelectAll,
  datatablesUnselectAll,
  renderSelectedAsInputsIntoElement,
} from "../utils";

const selected = new Set<string>();
const selectAllBtn = document.querySelector("#select-all-btn");
const unselectAllBtn = document.querySelector("#unselect-all-btn");
const deleteBtn = document.querySelector("#delete-btn");

function hideBtns(...els: Array<Element>) {
  for (const el of els) {
    el.classList.remove("dib");
    el.classList.add("dn");
  }
}
function showBtns(...els: Array<Element>) {
  for (const el of els) {
    el.classList.remove("dn");
    el.classList.add("dib");
  }
}
function redraw() {
  if (selected.size >= 1) {
    showBtns(deleteBtn);
  } else {
    hideBtns(deleteBtn);
  }
}

$("#table_id tbody").on("click", "tr", datatablesClickHandler(selected, redraw));
selectAllBtn.addEventListener("click", datatablesSelectAll(selected, redraw));
unselectAllBtn.addEventListener("click", datatablesUnselectAll(selected, redraw));
deleteBtn.addEventListener("click", function () {
  const deleteBtnForm = document.querySelector("form#delete-btn-form") as HTMLFormElement;
  const deleteBtnList = deleteBtnForm.querySelector("#delete-btn-list");
  renderSelectedAsInputsIntoElement(selected, deleteBtnList, "cohort");
  deleteBtnForm.submit();
});
