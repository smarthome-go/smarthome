import { LanguageSupport, LRLanguage } from '@codemirror/language';
declare const HomescriptLanguage: LRLanguage;
declare const HomescriptCompletion: import("@codemirror/state").Extension;
declare function Homescript(): LanguageSupport;
export { HomescriptLanguage, HomescriptCompletion, Homescript };
