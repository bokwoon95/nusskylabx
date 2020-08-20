import { Type, IQuestion } from "./formx.d";

export function isIQuestionArray(obj: any): obj is Array<IQuestion> {
  if (!Array.isArray(obj)) {
    return false;
  }
  for (const question of obj as Array<any>) {
    const hasCorrectType = Object.values(Type).includes(question.Type);
    const hasText = typeof question.Text === "string";
    const hasName = typeof question.Name === "string" || typeof question.Name === "undefined";
    const hasOptions = (function (options: any): boolean {
      if (options === null || options === undefined) {
        return true;
      }
      if (!Array.isArray(options)) {
        return false;
      }
      for (const option of options) {
        if (typeof option.Value !== "string") {
          return false;
        }
        if (typeof option.Display !== "string") {
          return false;
        }
      }
      return true;
    })(question.Options);
    const hasSubquestions = (function (subquestions: any): boolean {
      if (subquestions === null || subquestions === undefined) {
        return true;
      }
      if (!Array.isArray(subquestions)) {
        return false;
      }
      for (const subqn of subquestions) {
        if (typeof subqn.Name !== "string" && typeof subqn.Text !== "string") {
          return false;
        }
      }
      return true;
    })(question.Subquestions);
    if (!hasCorrectType || !hasText || !hasName || !hasOptions || !hasSubquestions) {
      return false;
    }
  }
  return true;
}
