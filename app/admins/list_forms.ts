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
const selectAllBtn = document.querySelector("#select-all");
const createFormBtn = document.querySelector("#create-form");
const deleteFormBtn = document.querySelector("#delete-form");
const unselectAllBtn = document.querySelector("#unselect-all");
const periodDuplicateBtn = document.querySelector("#period-duplicate");
const cohortDuplicateBtn = document.querySelector("#cohort-duplicate");

function redraw() {
  if (selected.size === 1) {
    showElement(deleteFormBtn);
    showElement(periodDuplicateBtn);
    showElement(cohortDuplicateBtn);
  } else if (selected.size > 1) {
    showElement(deleteFormBtn);
    hideElement(periodDuplicateBtn);
    showElement(cohortDuplicateBtn);
  } else {
    hideElement(deleteFormBtn);
    hideElement(periodDuplicateBtn);
    hideElement(cohortDuplicateBtn);
  }
}

$("#table_id tbody").on("click", "tr", datatablesClickHandler(selected, redraw));
selectAllBtn.addEventListener("click", datatablesSelectAll(selected, redraw));
unselectAllBtn.addEventListener("click", datatablesUnselectAll(selected, redraw));
createFormBtn.addEventListener("click", function () {
  MicroModal.show("create-form-form");
});
deleteFormBtn.addEventListener("click", function () {
  const deleteFormForm = document.querySelector("form#delete-form-form") as HTMLFormElement;
  const deleteFormFormInputs = deleteFormForm.querySelector("#delete-form-form-inputs");
  renderSelectedAsInputsIntoElement(selected, deleteFormFormInputs, "formID");
  deleteFormForm.submit();
});
periodDuplicateBtn.addEventListener("click", function () {
  MicroModal.show("duplicate-form-for-period");
  const formIDList = document.querySelector("#duplicate-form-for-period").querySelector(".formIDList");
  renderSelectedAsInputsIntoElement(selected, formIDList, "formID");
  redraw();
});
cohortDuplicateBtn.addEventListener("click", function () {
  MicroModal.show("duplicate-forms-for-cohort");
  const formIDList = document.querySelector("#duplicate-forms-for-cohort").querySelector(".formIDList");
  renderSelectedAsInputsIntoElement(selected, formIDList, "formID");
  redraw();
});
