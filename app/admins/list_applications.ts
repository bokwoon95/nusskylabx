import {
  // hideElement,
  // showElement,
  datatablesClickHandler,
  datatablesSelectAll,
  datatablesUnselectAll,
  // renderSelectedAsInputsIntoElement,
} from "../utils";
import MicroModal from "micromodal";
MicroModal.init();

const selected = new Set<string>();
const selectAllBtn = document.querySelector("#select-all-btn");
const unselectAllBtn = document.querySelector("#unselect-all-btn");

$("#table_id tbody").on("click", "tr", datatablesClickHandler(selected, null));
selectAllBtn.addEventListener("click", datatablesSelectAll(selected, null));
unselectAllBtn.addEventListener("click", datatablesUnselectAll(selected, null));
