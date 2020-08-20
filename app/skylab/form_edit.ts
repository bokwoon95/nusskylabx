import "regenerator-runtime/runtime"; // Needed for babel to compile async/await and generators to ES5
import { initFormbuilder } from "../../helpers/formx/formx";

const dataQuestions = document.getElementById("data-questions");
if (dataQuestions === null || dataQuestions === undefined) {
  throw new Error("#data-questions node not found");
}
const dataString = dataQuestions.dataset.questions.trim();
initFormbuilder({ selector: "#formx", data: dataString });
