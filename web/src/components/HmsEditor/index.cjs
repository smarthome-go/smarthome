'use strict';

Object.defineProperty(exports, '__esModule', { value: true });

var lr = require('@lezer/lr');
var language = require('@codemirror/language');
var highlight = require('@lezer/highlight');

// This file was generated by lezer-generator. You probably shouldn't edit it.
const parser = lr.LRParser.deserialize({
  version: 14,
  states: "nQYQPOOOOQO'#Ch'#ChOOQO'#Cd'#CdQYQPOOOOQO-E6b-E6b",
  stateData: "k~OZOSPOS~ORPOSPOTPOUPOVPO~O",
  goto: "h]PPPPPPPP^PPPdQRORSRTQOR",
  nodeNames: "⚠ LineComment Program Boolean Identifier Keyword String Number",
  maxTerm: 12,
  skippedNodes: [0,1],
  repeatNodeCount: 1,
  tokenData: "5P~ReXY!dYZ!d]^!dpq!drs!ust#iwx#t!Q![$c!c!}$|#T#X$|#X#Y%}#Y#Z)Y#Z#]$|#]#^,_#^#b$|#b#c/T#c#d0}#d#g$|#g#h2c#h#i3q#i#o$|~!iSZ~XY!dYZ!d]^!dpq!d~!zUU~OY!uZr!urs#^s#O!u#O#P#c#P~!u~#cOU~~#fPO~!u~#nQP~OY#iZ~#i~#yUU~OY#tZw#twx#^x#O#t#O#P$]#P~#t~$`PO~#t~$hQV~!O!P$n!Q![$c~$qP!Q![$t~$yPV~!Q![$t~%PVXY%fYZ%f]^%fpq%fxy%x!c!}$|#T#o$|~%iTXY%fYZ%f]^%fpq%fxy%x~%}OS~~&QXXY%fYZ%f]^%fpq%fxy%x!c!}$|#T#`$|#`#a&m#a#o$|~&pXXY%fYZ%f]^%fpq%fxy%x!c!}$|#T#g$|#g#h']#h#o$|~'`XXY%fYZ%f]^%fpq%fxy%x!c!}$|#T#X$|#X#Y'{#Y#o$|~(OWXY(hYZ(h]^(hpq(hxy%x!c!}$|#T#o$|#o#p(}~(kUXY(hYZ(h]^(hpq(hxy%x#o#p(}~)QP#q#r)T~)YOT~~)]WXY%fYZ%f]^%fpq%fxy%x!c!}$|#T#U)u#U#o$|~)xXXY%fYZ%f]^%fpq%fxy%x!c!}$|#T#`$|#`#a*e#a#o$|~*hXXY%fYZ%f]^%fpq%fxy%x!c!}$|#T#g$|#g#h+T#h#o$|~+WXXY%fYZ%f]^%fpq%fxy%x!c!}$|#T#X$|#X#Y+s#Y#o$|~+xVR~XY%fYZ%f]^%fpq%fxy%x!c!}$|#T#o$|~,bXXY%fYZ%f]^%fpq%fxy%x!c!}$|#T#Y$|#Y#Z,}#Z#o$|~-QWXY-jYZ-j]^-jpq-jxy%x!c!}$|#T#o$|#o#p.P~-mUXY-jYZ-j]^-jpq-jxy%x#o#p.P~.SP#q#r.V~.[PT~#X#Y._~.bP#`#a.e~.hP#g#h.k~.nP#X#Y.q~.tTXY.qYZ.q]^.qpq.q#o#p(}~/WXXY%fYZ%f]^%fpq%fxy%x!c!}$|#T#i$|#i#j/s#j#o$|~/vXXY%fYZ%f]^%fpq%fxy%x!c!}$|#T#a$|#a#b0c#b#o$|~0hVT~XY%fYZ%f]^%fpq%fxy%x!c!}$|#T#o$|~1QZXY%fYZ%f]^%fpq%fxy%x!c!}$|#T#Y$|#Y#Z1s#Z#b$|#b#c+s#c#o$|~1vXXY%fYZ%f]^%fpq%fxy%x!c!}$|#T#Y$|#Y#Z+s#Z#o$|~2fXXY%fYZ%f]^%fpq%fxy%x!c!}$|#T#h$|#h#i3R#i#o$|~3UXXY%fYZ%f]^%fpq%fxy%x!c!}$|#T#f$|#f#g0c#g#o$|~3tXXY%fYZ%f]^%fpq%fxy%x!c!}$|#T#f$|#f#g4a#g#o$|~4dXXY%fYZ%f]^%fpq%fxy%x!c!}$|#T#i$|#i#j+T#j#o$|",
  tokenizers: [0],
  topRules: {"Program":[0,2]},
  tokenPrec: 0
});

const HomescriptLanguage = language.LRLanguage.define({
    parser: parser.configure({
        props: [
            highlight.styleTags({
                // Identifier: t.constant,
                Keyword: highlight.tags.keyword,
                Boolean: highlight.tags.bool,
                String: highlight.tags.string,
                Number: highlight.tags.integer,
                LineComment: highlight.tags.lineComment,
                "( )": highlight.tags.paren
            }),
            language.indentNodeProp.add({
                Application: context => context.column(context.node.from) + context.unit
            }),
            language.foldNodeProp.add({
                Application: language.foldInside
            })
        ]
    }),
    languageData: {
        commentTokens: { line: "#" }
    }
});
function Homescript() {
    return new language.LanguageSupport(HomescriptLanguage);
}

exports.Homescript = Homescript;
exports.HomescriptLanguage = HomescriptLanguage;
