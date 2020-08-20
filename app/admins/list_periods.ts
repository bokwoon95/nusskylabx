import {
  hideElement,
  showElement,
  datatablesClickHandler,
  datatablesSelectAll,
  datatablesUnselectAll,
  renderSelectedAsInputsIntoElement,
} from "../utils";
import MicroModal from "micromodal";
MicroModal.init();

const selected = new Set<string>();
const selectAllBtn = document.querySelector("#select-all-btn");
const unselectAllBtn = document.querySelector("#unselect-all-btn");
const createBtn = document.querySelector("#create-btn");
const deleteBtn = document.querySelector("#delete-btn");
const cohortDuplicateBtn = document.querySelector("#cohort-duplicate-btn");

function redraw() {
  if (selected.size === 1) {
    showElement(deleteBtn);
    showElement(cohortDuplicateBtn);
  } else if (selected.size > 1) {
    showElement(deleteBtn);
    showElement(cohortDuplicateBtn);
  } else {
    hideElement(deleteBtn);
    hideElement(cohortDuplicateBtn);
  }
}

$("#table_id tbody").on("click", "tr", datatablesClickHandler(selected, redraw));
selectAllBtn.addEventListener("click", datatablesSelectAll(selected, redraw));
unselectAllBtn.addEventListener("click", datatablesUnselectAll(selected, redraw));
createBtn.addEventListener("click", function () {
  MicroModal.show("create-btn-form");
});
deleteBtn.addEventListener("click", function () {
  const deleteBtnForm = document.querySelector("form#delete-btn-form") as HTMLFormElement;
  const deleteBtnList = deleteBtnForm.querySelector("#delete-btn-list");
  renderSelectedAsInputsIntoElement(selected, deleteBtnList, "periodID");
  deleteBtnForm.submit();
});
cohortDuplicateBtn.addEventListener("click", function () {
  MicroModal.show("cohort-duplicate-form");
  const cohortDuplicateForm = document.querySelector("form#cohort-duplicate-form") as HTMLFormElement;
  const cohortDuplicateList = cohortDuplicateForm.querySelector("#cohort-duplicate-list");
  renderSelectedAsInputsIntoElement(selected, cohortDuplicateList, "periodID");
  redraw();
});
