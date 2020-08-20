export type EventHandler = (event: Event) => void;

export enum Type {
  Paragraph = "paragraph",
  Shorttext = "short text",
  Longtext = "long text",
  Checkbox = "checkbox",
  Select = "select",
  Radio = "radio",
  Multiradio = "multiradio",
  Date = "date",
  Time = "time",
  Image = "image",
  // Null = "",
}

export enum FormStatus {
  ViewForm = "viewform",
  ViewJSON = "viewjson",
  LoadJSON = "loadjson",
}

export interface IOption {
  Value: string;
  Display: string;
}

export interface ISubquestion {
  Name: string;
  Text: string;
}

export interface IQuestion {
  Type: Type;
  Text: string;
  Name?: string;
  Options?: Array<IOption>;
  Subquestions?: Array<ISubquestion>;
}

export interface ISubquestionAnswer {
  Name: string;
  Text: string;
  Answer?: Array<string>;
}

export interface IQuestionAnswer {
  Type: Type;
  Text: string;
  Name?: string;
  Options?: Array<IOption>;
  Subquestions?: Array<ISubquestionAnswer>;
  Answer?: Array<string>;
}

export interface INode {
  uuid: string;
  question: IQuestion;
}
