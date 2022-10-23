import { LRLanguage, indentNodeProp, delimitedIndent, foldNodeProp, foldInside, LanguageSupport } from '@codemirror/language';
import { styleTags, tags } from '@lezer/highlight';
import { LRParser } from '@lezer/lr';

// This file was generated by lezer-generator. You probably shouldn't edit it.
const spec_Ident = {__proto__:null,let:10, true:16, false:16, on:16, off:16, null:22, exit:28, throw:28, assert:28, print:28, sleep:28, switch_on:28, switch:28, notify:28, log:28, exec:28, get:28, http:28, user:32, weather:32, time:32, if:36, else:46, for:50, in:52, while:56, loop:60, fn:64, try:76, catch:78, as:100, str:104, bool:104, num:104, import:108, from:110, break:114, continue:118, return:122};
const parser = LRParser.deserialize({
  version: 14,
  states: "0bQYQPOOOOQO'#Cc'#CcO!pQQO'#DuOOQO'#Ci'#CiOOQO'#Ck'#CkOOQO'#Ch'#ChO$sQPO'#DUO$sQPO'#DVO'hQQO'#DsOOQO'#Du'#DuOOQO'#Ds'#DsO'rQPO'#DlQYQPOOO$sQPO'#CmO'wQPO'#CtO$sQPO'#CwO'|QPO'#CyO(RQPO'#C{O'|QPO'#DRO(ZQPO'#C_O(`QPO'#DcO(eQPO'#DfOOQO'#Dh'#DhO(oQPO'#DjO$sQPO,59QO(yQQO,59pO)QQQO,59qO$sQPO,59rO$sQPO,59rO$sQPO,59rO$sQPO,59rO$sQPO,59rO$sQPO,59rO$sQPO,59rO+QQSO,59sO+VQPO'#D]OOQO,59v,59vO$sQPO,59xO+^QPO,59yOOQO,5:W,5:WOOQO-E7j-E7jO+fQQO,59XO+mQPO,59`O+fQQO,59cO+rQPO'#CoOOQO,59e,59eO+yQPO'#C}O'|QPO,59gO(UQPO,59gO,RQPO,59mO,WQPO,58yO,cQPO,59}O,kQQO,5:QO,uQQO,5:UOOQO1G.l1G.lOOQO1G/[1G/[O.|QQO1G/^O/TQQO1G/^O1OQQO1G/^O1`QQO1G/^O3gQQO1G/^O3tQQO1G/^O4RQQO1G/^OOQO'#DZ'#DZOOQO1G/_1G/_O6RQQO,59wOOQO,59w,59wO6]QQO1G/dOOQO'#Da'#DaOOQO1G/e1G/eO6sQQO1G.sO$sQPO1G.zOOQO1G.}1G.}OOQO,59Z,59ZO8vQPO,59ZOOQO,59i,59iO9OQPO,59iOOQO1G/R1G/RO'|QPO1G/RO9WQPO1G/XO$sQPO1G.eO9]QPO1G/iO9bQPO1G/iO$sQPO'#DoO9gQPO1G/cOOQO1G/c1G/cO9oQPO7+$_O9wQQO7+$fO:OQPO1G.uO:VQPO1G.uOOQO1G.u1G.uO:_QPO'#DnO:dQPO1G/TOOQO1G/T1G/TOOQO7+$m7+$mO'|QPO7+$sO:lQQO7+$POOQO7+%T7+%TO:vQPO7+%TO:{QQO,5:ZOOQO-E7m-E7mOOQO7+$}7+$}OOQO<<Gy<<GyO$sQPO<<HQOOQO,5:X,5:XOOQO7+$a7+$aO;VQPO7+$aOOQO-E7k-E7kOOQO,5:Y,5:YOOQO-E7l-E7lOOQO7+$o7+$oOOQO<<H_<<H_O;^QPO<<HoO+fQQOAN=lOOQO<<G{<<G{PYQPO'#DmOOQOAN>ZAN>ZOOQOG23WG23W",
  stateData: ";c~O!fOSPOS~OSTOTcOUXOWPOXQOZXO^RO`SOb]Oi^Ol_On`OpaOrUOvbO!WdO!ZeO!]fO!_gO!lVO!mVO!nVO~O!jhOe!iXr!iX|!iX!S!iX!h!iX!l!iX!m!iX!o!iX!p!iX!q!iX!r!iX!s!iX!t!iX!u!iX!v!iX!w!iX!x!iX!y!iX!z!iX!|!iX!}!iX#O!iX#P!iX#Q!iX#R!iXt!iXd!iXs!iXf!iX!k!iX~OSTOUXOWPOXQOZXO^RO`SOb]Oi^Ol_On`OpaOrUOvbO!lVO!mVO!nVO~OrsO|rO!SvO!huO!loO!moO!okO!plO!qmO!rmO!snO!tnO!unO!vnO!wpO!xpO!ypO!zqO!|uO!}uO#OuO#PuO#QuO#RuO~Oe!gXf!gX~P%zOewO~OSzO~Od|O~OS!QOr!OO~OS!SO~OS!TO~Oe!YXf!YX~P$sOe!^Xf!^X~P$sOt!XO~P%zOrsO|rOeya!Sya!hya!lya!mya!oya!pya!qya!rya!sya!tya!uya!vya!wya!xya!yya!zya!|ya!}ya#Oya#Pya#Qya#Ryatyadyasyafya!kya~O!{!aO~Ot!dO~P$sOZ!fO!U!fO~Od|O~P%zOj!iO~Of!kO~PYOS!nOt!mO~Ow!qO~O!h!rOeRafRa~O!S!tO!X!sO~Oe!Yaf!Ya~P%zOe!^af!^a~P%zOrsO|rO!SvO!loO!moO!qmO!rmO!snO!tnO!unO!vnO!wpO!xpO!ypO!zqOezi!hzi!ozi!|zi!}zi#Ozi#Pzi#Qzi#Rzitzidziszifzi!kzi~O!plO~P-PO!pzi~P-POrsO|rO!SvO!loO!moO!wpO!xpO!ypO!zqOezi!hzi!ozi!pzi!qzi!rzi!|zi!}zi#Ozi#Pzi#Qzi#Rzitzidziszifzi!kzi~O!snO!tnO!unO!vnO~P/[O!szi!tzi!uzi!vzi~P/[OrsO|rO!SvO!zqOezi!hzi!lzi!mzi!ozi!pzi!qzi!rzi!szi!tzi!uzi!vzi!|zi!}zi#Ozi#Pzi#Qzi#Rzitzidziszifzi!kzi~O!wpO!xpO!ypO~P1pO!wzi!xzi!yzi~P1pOrsO|rO!zqOezi!Szi!hzi!lzi!mzi!ozi!pzi!qzi!rzi!szi!tzi!uzi!vzi!wzi!xzi!yzi!|zi!}zi#Ozi#Pzi#Qzi#Rzitzidziszifzi!kzi~Os!uOt!wO~P%zOe!Qit!Qid!Qis!Qif!Qi!k!Qi~P%zOg!xOeairai|ai!Sai!hai!lai!mai!oai!pai!qai!rai!sai!tai!uai!vai!wai!xai!yai!zai!|ai!}ai#Oai#Pai#Qai#Raitaidaisaifai!kai~Oe!zOf!|O~Os!}Ot#PO~OS#RO~OS#TO~OS#UO~Os!uOt#XO~Ob]Od|O~O!k#ZO~P%zOf#]O~PYOe#^Of#]O~OS#`O~Os!}Ot#bO~OeRqfRq~P%zO!X#dO~Os!cat!ca~P%zOf#fO~PYOS#hO~O",
  goto: "(Z!jPPP!kPPP!sPP!sP!s#aP#aP#}P$nPPPP!sPP!sP!sP!sP%ZPPP!sPP!s!s!s!sP%a!s%d!s!sP%zP!kPP!kP!kP!kP%}&T&Z&aPPP&gP&s]YO[|!z#^#g!VXOUV[]_eghklmnopqsu|!i!r!u!z#Z#^#g!VTOUV[]_eghklmnopqsu|!i!r!u!z#Z#^#g!UXOUV[]_eghklmnopqsu|!i!r!u!z#Z#^#gR#Y!xQ}`Q!RbQ!hyQ!j{Q!o!PQ#Q!pQ#Y!xQ#c#RR#i#eQ!PaR!p!QR!brytWijy{!U!V!Y!Z![!]!^!_!`!c!e!y#S#V#eR!gvQ[ORx[Q!{!lR#_!{Q#O!nR#a#OQ!v!cR#W!vSZO[Q!l|V#[!z#^#g[WO[|!z#^#gQiUQjVQy]Q{_Q!UeQ!VgQ!WhQ!YkQ!ZlQ![mQ!]nQ!^oQ!_pQ!`qQ!csQ!euQ!y!iQ#S!rQ#V!uR#e#Z",
  nodeNames: "⚠ Comment Program LetStmt Ident let Number Bool Bool String PairExpr null VariableName BuiltinFunc BuiltinFunc BuiltinVar BuiltinVar IfExpr if Block { ; } else ForExpr for in WhileExpr while LoopExpr loop FnExpr fn Parameters ( , ) TryExpr try catch ParenExpr PrefixExpr InfixExpr MemberExpr . Property CallExpr Arguments AssignExpr CastExpr as Type Type ImportStmt import from BreakStmt break ContinueStmt continue ReturnStmt return",
  maxTerm: 95,
  skippedNodes: [0,1],
  repeatNodeCount: 4,
  tokenData: ")}~RmXY!|YZ!|]^!|pq!|qr#Rrs#`st$Puv$[vw$iwx$txy%`yz%ez{%j{|&X|}&f}!O&k!O!P&x!P!Q'V!Q!['d!]!^'}!^!_(S!_!`(a!`!a(v!c!})T#R#S)T#T#o)T#o#p)h#p#q)m#q#r)x~#RO!f~R#WP!nP!_!`#ZQ#`O!rQ~#eTX~Or#`rs#ts#O#`#O#P#y#P~#`~#yOX~~#|PO~#`~$UQP~OY$PZ~$P~$aP!y~!_!`$d~$iO#O~~$lPvw$o~$tO!p~~$yTX~Ow$twx#tx#O$t#O#P%Y#P~$t~%]PO~$t~%eOr~~%jOt~~%oQ!w~z{%u!_!`&S~%zP!z~!_!`%}~&SO#R~~&XO!|~~&^P!l~!_!`&a~&fO#P~~&kOs~~&pP!m~!_!`&s~&xO#Q~~&}P|~!O!P'Q~'VO!k~~'[P!x~!_!`'_~'dO!}~~'iQU~!O!P'o!Q!['d~'rP!Q!['u~'zPU~!Q!['u~(SOe~~(XP!s~!_!`([~(aO!t~~(fQ!h~!_!`(l!`!a(q~(qO!q~~(vO!j~~({P!u~!_!`)O~)TO!v~V)[SSR!{S!Q![)T!c!})T#R#S)T#T#o)T~)mOd~~)pP#p#q)s~)xO!o~~)}Of~",
  tokenizers: [0, 1, 2],
  topRules: {"Program":[0,2]},
  specialized: [{term: 4, get: value => spec_Ident[value] || -1}],
  tokenPrec: 0
});

const HomescriptLanguage = LRLanguage.define({
    parser: parser.configure({
        props: [
            indentNodeProp.add({
                Application: delimitedIndent({ closing: ')', align: false }),
            }),
            foldNodeProp.add({
                Application: foldInside,
            }),
            styleTags({
                'for while loop if else try catch return break continue': tags.controlKeyword,
                in: tags.operatorKeyword,
                'let fn': tags.definitionKeyword,
                'import from': tags.moduleKeyword,
                as: tags.keyword,
                Bool: tags.bool,
                null: tags.null,
                Type: tags.typeName,
                'VariableName/Ident': tags.variableName,
                'CallExpr/VariableName/Ident': tags.function(tags.variableName),
                Property: tags.propertyName,
                'CallExpr/MemberExpr/Property': tags.function(tags.propertyName),
                'FnExpr/Ident': tags.function(tags.variableName),
                'Parameters/Ident': tags.local(tags.variableName),
                Comment: tags.lineComment,
                Number: tags.number,
                String: tags.string,
                '+ - "*" "/" % "**"': tags.arithmeticOperator,
                '|| &&': tags.logicOperator,
                '< <= > >= "!=" ==': tags.compareOperator,
                '=': tags.definitionOperator,
                '( ) { }': tags.bracket,
                '. , ;': tags.separator,
                BuiltinFunc: tags.standard(tags.function(tags.variableName)),
                BuiltinVar: tags.standard(tags.variableName),
            }),
        ],
    }),
    languageData: {
        commentTokens: { line: '#' },
    },
});
function Homescript() {
    return new LanguageSupport(HomescriptLanguage);
}

export { Homescript, HomescriptLanguage };
