'use strict';

Object.defineProperty(exports, '__esModule', { value: true });

var language = require('@codemirror/language');
var autocomplete = require('@codemirror/autocomplete');
var highlight = require('@lezer/highlight');
var lr = require('@lezer/lr');

// This file was generated by lezer-generator. You probably shouldn't edit it.
const spec_Ident = {__proto__:null,import:14, type:18, templ:22, trigger:24, from:32, let:40, return:66, break:70, continue:74, loop:78, while:82, for:86, in:88, true:100, false:100, on:100, off:100, new:106, fn:110, null:120, none:122, exit:132, throw:132, assert:132, print:132, println:132, debug:132, as:196, if:200, else:202, match:206, try:212, catch:214, pub:220};
const parser = lr.LRParser.deserialize({
  version: 14,
  states: "BbQ]QPOOOOQO'#C`'#C`OOQO'#Eo'#EoQ]QPOOOkQPO'#CaO|QPO'#CoO!RQPO'#ElO!WQPO'#EmOOQO-E8m-E8mOOQO'#Cd'#CdO!cQPO'#CdO!hQPO'#CdO!mQPO,58{O!rQPO,58{O#QQPO,59ZO#YQPO,5;WO#_QPO'#C{OOQO,5;X,5;XOOQO'#Cf'#CfOOQO,59O,59OO#dQPO1G.gO#iQPO1G.gO#qQPO1G.uO$xQQO1G.uO%WQPO'#DfO%`QQO1G0rO%hQPO,59gOOQO'#Cm'#CmO%mQPO7+$RO%rQPO7+$RO&TQPO7+$RO&]QPO7+$RO&bQPO'#DaO(YQSO7+$aO(aQPO'#D]OOQO'#D]'#D]OOQO'#Do'#DoOOQO'#Dn'#DnOOQO'#Dr'#DrO#qQPO'#DqO(fQPO'#E[OOQO'#D['#D[O#qQPO'#D[O*[QPO'#CyOOQO'#Cx'#CxOOQO'#Cw'#CwO*cQPO'#DbO#YQPO'#DdO#qQPO'#EcO#qQPO'#EfO*hQPO'#EiO$xQQO'#E{O*mQSO'#E|OOQO'#Cq'#CqO$xQQO'#CqO*{QPO7+$aOOQO,5:Q,5:QO+QQPO,5:QOOQO7+&^7+&^O$xQQO7+&^O$xQQO1G/ROOQO<<Gm<<GmOOQO,5;[,5;[O+VQPO<<GmO+[QPO<<GmOOQO-E8n-E8nO#dQPO<<GmOOQO,59{,59{O+mQSO,59{O#qQPO,5:WO#qQPO,5:aO#qQPO,5:aO#qQPO,5:aO#qQPO,5:aO#qQPO,5:aO#qQPO,5:aO#qQPO,5:aO#qQPO,5:aO#qQPO,5:aO#qQPO,5:tO#qQPO,5:xO+wQPO,5:yO$xQQO,5:{OOQO<<G{<<G{OOQO'#D_'#D_OOQO,59w,59wO+|QSO,5:]O.SQPO,5:vO.ZQSO,59vO.bQSO,59eO0YQSO'#CwO1zQWO'#DZOOQO'#Cz'#CzOOQO'#Er'#ErO3sQPO,59eOOQO,59e,59eO3zQPO'#C|O4RQPO'#DOO4WQPO'#DQO*hQPO'#DSO#qQPO'#DUO4]QPO'#DWO4bQSO,59|O4pQQO,5:OO4xQSO,5:}O5PQSO,5;QO5WQPO,5;TO5]QPO,5;gO5bQPO'#FPOOQO'#E}'#E}O5gQPO'#E}OOQO,5;h,5;hO5oQPO,5;hOOQO,59],59]O#qQPO<<G{O$xQQO1G/lO*hQPO<<IxO5tQPO7+$mO#dQPOAN=XO5yQPOAN=XP!rQPO'#EpO6OQPOAN=XO6TQPO1G/gO6[QPO1G/gOOQO1G/g1G/gO6dQSO1G/rO:mQSO1G/{O:tQSO1G/{O<{QSO1G/{O=VQSO1G/{O?^QSO1G/{O?kQSO1G/{OA{QSO1G/{OBSQSO1G/{OC[QSO1G/{ODeQSO1G0`OEnQSO1G0dOOQO1G0e1G0eOOQO1G0g1G0gOEuQSO'#FTOOQO1G0b1G0bOFPQPO1G0bOOQO1G/b1G/bOOQO1G/P1G/POOQO,59u,59uOFUQSO1G/POOQO-E8p-E8pOF]QSO,59hOOQO,59h,59hOOQO,59j,59jOOQO,59l,59lOFdQPO,59nO4xQSO,59pOHZQPO,59rOH`QPO'#FSOOQO'#FR'#FROHeQPO'#FROOQO1G/h1G/hOHmQPO1G/hOOQO1G/j1G/jO$xQQO1G/jOJlQSO1G0iOLWQPO1G0lOL_QPO1G0oOOQO1G1R1G1RO$xQQO,5;kOLdQPO,5;iOLoQPO,5;iOOQO1G1S1G1SOLwQSOAN=gOMOQPO7+%WOOQOAN?dAN?dOOQO<<HX<<HXOMWQPOG22sO#dQPOG22sOOQOG22sG22sOM]QSO,5;_OOQO7+%R7+%ROMjQPO7+%ROOQO-E8q-E8qOOQO7+&O7+&OOMqQPO,5;oOMxQPO,5;oOOQO7+%|7+%|OOQO7+$k7+$kO&iQSOOOOQO1G/S1G/SOOQO1G/Y1G/YONQQPO1G/[O#qQPO1G/^O! wQWO1G0iO#qQPO,5;nO!#aQPO,5;mO!#lQPO,5;mOOQO7+%S7+%SO*hQPO7+%UO!#tQPO7+&TO!#|QSO'#DnO#qQPO'#DqO!'aQSO'#D[OOQO7+&W7+&WO!'hQPO7+&WO*hQPO7+&ZOOQO1G1V1G1VOOQO,5;],5;]O!'pQPO1G1TOOQO-E8o-E8oOOQOG23RG23RO!'{QPO'#EuO!(QQPO<<HrOOQO<<Hr<<HrOOQOLD(_LD(_O!(YQPOLD(_OOQO<<Hm<<HmP#qQPO'#EsO!(_QPO1G1ZOOQO7+$v7+$vO4xQSO7+$xO!(fQSO1G1YOOQO,5;`,5;`O!(pQPO1G1XOOQO-E8r-E8rOOQO<<Hp<<HpOOQO<<Io<<IoO#qQPO,5;pO!({QSO,5;pO!)SQPO<<IrO!)ZQPO<<IrOOQO<<Ir<<IrOOQO<<Iu<<IuP!)cQPO'#EqO!)kQPO,5;aOOQO-E8s-E8sOOQOAN>^AN>^OOQO!$'Ky!$'KyO!)pQPO<<HdP!+gQPO'#EtO!+oQSO1G1[O!+vQSO1G1[O#qQPO1G1[OOQO,5;b,5;bOOQOAN?^AN?^O!,QQPOAN?^OOQO-E8t-E8tO$xQQO1G0{OOQOAN>OAN>OOOQO7+&v7+&vO!,XQSO7+&vO!,`QSO7+&vOOQOG24xG24xPKPQPO'#EvOOQO7+&g7+&gOOQO<<Jb<<JbO4xQSO,5:}O#qQPO'#EcO!,jQPO7+&T",
  stateData: "!,u~O#mOSPOSQOS~OVSOdTO!XUO#bVO~OUXOXYOZZO[ZO]]O~OU^O~OU_O~OX`OdTO!XUO~OUbO~OUcO~O`dO~OUXOXYOZZO[ZO~O#ngO#tfO~O!ZhO~OUjO~OUkO~O^mO_oO~OUuO]{OgpOisOjvO!QrO!V!OO!X!PO!ZzO!^sO!_sO!dtO!gvO!hvO#W!QO#Z!RO#^!SO~O]!UOf!VOg!TOj!WO~OU!ZO![!YO~O]{O!]!]O~O#t!^O~Ob!_O~OUXOXYOZZO[ZO_!aO~O^!bO_!aO~O`!dO~Oh!eO~P#qOg!rO!a!gO!g!lO!j!hO!k!iO!l!jO!m!kO!n!kO!o!lO!p!mO!q!mO!r!mO!s!nO!t!oO!u!nO!v!pO!w!pO!x!pO!y!pO!z!pO!{!pO!}!qO#S!sO#U!tO!Z#PX~Ob!uO~P&iO!S!vO~O!Z!yO~OUuOX`O]{OdTOgpOisOjvOq#SOs#TOu#UOw#VOy#WO{#XO!QrO!V!OO!X!PO!ZzO!^sO!_sO!dtO!gvO!hvO#W&hO#Z!RO#^!SO~O_#RO~P(kO]#YO~O]{O~OU#`O_#cOi#`O#r#aO~O#t#fO~O#n#gO~O`#jO~OUXOXYOZZO[ZO_#kO~O^#nOh#pO~P&iOU#}O~O!s!nO#S!sOb!eag!ea!Z!ea!a!ea!g!ea!j!ea!k!ea!l!ea!m!ea!n!ea!o!ea!p!ea!q!ea!r!ea!t!ea!u!ea!v!ea!w!ea!x!ea!y!ea!z!ea!{!ea!}!ea#U!ea^!eah!ea![!ea_!ea]!ea#[!ea~O![$QO~P#qO![$SO~P&iO_$TO~P&iOgkX!ZkX!akX!gkX!jkX!kkX!lkX!mkX!nkX!okX!pkX!qkX!rkX!skX!tkX!ukX!vkX!wkX!xkX!ykX!zkX!{kX!}kX#SkX#UkX~Ob$UO_kX~P.iO!akX!jkX!kkX!lkX!mkX!nkX!okX!pkX!qkX!rkX!skX!tkX!ukX!vkX!wkX!xkX!ykX!zkX!{kX!}kX#SkX#UkX~Ob$UOU}XX}X]}X_}Xd}Xg}Xi}Xj}Xq}Xs}Xu}Xw}Xy}X{}X!Q}X!V}X!X}X!Z}X!^}X!_}X!d}X!g}X!h}X#W}X#Z}X#^}X~P0dO_$TO~P(kOb$YO~P#qOb$ZO~Ob$[O~OU$_O~OU$`O_$cOi$`O#r$aO~O]{O!]$fO~O]{O~P&iO]$hO~P&iO#_$iO~Oh$jO~O#n$kO~O^$lO_#qX~O_$nO~Ob$rO~O`$tO~Ob$uO~Oh$wO~P#qO^$xOh$wO~O!s!nO!t!oO!u!nO#S!sOb!`ig!`i!Z!`i!a!`i!g!`i!j!`i!k!`i!l!`i!m!`i!n!`i!o!`i!p!`i!q!`i!r!`i!v!`i!w!`i!x!`i!y!`i!z!`i!{!`i!}!`i#U!`i^!`ih!`i![!`i_!`i]!`i#[!`i~O!a!gO!g!lO!l!jO!m!kO!n!kO!o!lO!p!mO!q!mO!r!mO!s!nO!t!oO!u!nO#S!sOb!iig!ii!Z!ii!j!ii!v!ii!w!ii!x!ii!y!ii!z!ii!{!ii!}!ii#U!ii^!iih!ii![!ii_!ii]!ii#[!ii~O!k!iO~P8jO!k!ii~P8jO!a!gO!g!lO!o!lO!p!mO!q!mO!r!mO!s!nO!t!oO!u!nO#S!sOb!iig!ii!Z!ii!j!ii!k!ii!l!ii!v!ii!w!ii!x!ii!y!ii!z!ii!{!ii!}!ii#U!ii^!iih!ii![!ii_!ii]!ii#[!ii~O!m!kO!n!kO~P:{O!m!ii!n!ii~P:{O!a!gO!s!nO!t!oO!u!nO#S!sOb!iig!ii!Z!ii!g!ii!j!ii!k!ii!l!ii!m!ii!n!ii!o!ii!v!ii!w!ii!x!ii!y!ii!z!ii!{!ii!}!ii#U!ii^!iih!ii![!ii_!ii]!ii#[!ii~O!p!mO!q!mO!r!mO~P=aO!p!ii!q!ii!r!ii~P=aO!s!nO#S!sOb!iig!ii!Z!ii!a!ii!g!ii!j!ii!k!ii!l!ii!m!ii!n!ii!o!ii!p!ii!q!ii!r!ii!t!ii!v!ii!w!ii!x!ii!y!ii!z!ii!{!ii!}!ii#U!ii^!iih!ii![!ii_!ii]!ii#[!ii~O!u!ii~P?xO!u!nO~P?xO!a!gO!g!lO!j!hO!k!iO!l!jO!m!kO!n!kO!o!lO!p!mO!q!mO!r!mO!s!nO!t!oO!u!nO#S!sO~Ob!iig!ii!Z!ii!v!ii!w!ii!x!ii!y!ii!z!ii!{!ii!}!ii#U!ii^!iih!ii![!ii_!ii]!ii#[!ii~PBZO!v!pO!w!pO!x!pO!y!pO!z!pO!{!pO!}!qOb!|ig!|i!Z!|i#U!|i^!|ih!|i![!|i_!|i]!|i#[!|i~PBZOh$zO~P&iO^${O![#wX~P&iO![$}O~O_%OO~P&iOb%QO~P&iOb%ROUvaXva]va_vadvagvaivajvaqvasvauvawvayva{va!Qva!Vva!Xva!Zva!^va!_va!dva!gva!hva#Wva#Zva#^va~O|%TO~O#n%VO~O^%WO_#uX~O_%YO~Ob#Vig#Vi!Z#Vi!a#Vi!g#Vi!j#Vi!k#Vi!l#Vi!m#Vi!n#Vi!o#Vi!p#Vi!q#Vi!r#Vi!s#Vi!t#Vi!u#Vi!v#Vi!w#Vi!x#Vi!y#Vi!z#Vi!{#Vi!}#Vi#S#Vi#U#Vi]#Vi_#Vi~O#X%[O^#Vih#Vi![#Vi#[#Vi~PHrOU%]O]{OgpOisOjvO!QrO!V!OO!X!PO!ZzO!^sO!_sO!dtO!gvO!hvO#W!QO#Z!RO#^!SO~O_%`O~PKPOU%bO~OU#`Oi#`O_#qa~O^%eO_#qa~Ob%gO~P&iO^%hO![%jO~Ob%kO~O^#gah#ga![#ga~P&iOh%mO~P#qO![#wa~P#qO^%oO![#wa~Ob%pOUxiXxi]xi_xidxigxiixijxiqxisxiuxiwxiyxi{xi!Qxi!Vxi!Xxi!Zxi!^xi!_xi!dxi!gxi!hxi#Wxi#Zxi#^xi~O#X&iOU#ViX#Vid#Vii#Vij#Viq#Vis#Viu#Viw#Viy#Vi{#Vi!Q#Vi!V#Vi!X#Vi!^#Vi!_#Vi!d#Vi!h#Vi#W#Vi#Z#Vi#^#Vi~PHrOU$`Oi$`O_#ua~O^%tO_#ua~O]{O#W!QO~O#[%xOg!bX!Z!bX!a!bX!g!bX!j!bX!k!bX!l!bX!m!bX!n!bX!o!bX!p!bX!q!bX!r!bX!s!bX!t!bX!u!bX!v!bX!w!bX!x!bX!y!bX!z!bX!{!bX!}!bX#S!bX#U!bX~Og!OX!Z!OX!a!OX!g!OX!j!OX!k!OX!l!OX!m!OX!n!OX!o!OX!p!OX!q!OX!r!OX!s!OX!t!OX!u!OX!v!OX!w!OX!x!OX!y!OX!z!OX!{!OX!}!OX#S!OX#U!OX~O#[%xO~P!%pO^%zO_%|O~OU#`Oi#`O_#qi~OU&PO~O^%hO![&RO~Ob&SO~O![#wi~P#qO^#vi_#vi~P&iOU$`Oi$`O_#ui~O#[&XO~P!%pO_&ZO~PKPO^&[O_&ZO~OU#`Oi#`O~O#n&^O~Ob&_OUzyXzy]zy_zydzygzyizyjzyqzyszyuzywzyyzy{zy!Qzy!Vzy!Xzy!Zzy!^zy!_zy!dzy!gzy!hzy#Wzy#Zzy#^zy~OU$`Oi$`O~O^&`O~P.iO^#xi_#xi~P.iO_&cO~PKPO^&fO~P.iO^#xq_#xq~P.iO]{O#W&hO~OP!q~",
  goto: "5]#yPPPP#z$OPP$SP$_PPPPPP$bP$nP$yPPPPP%i'n(o*a$r$uP$uP$uP$uP$uP$uPP$u*e+fP,eP,h,hP,hP-cPPPPP,hP-i.dP-i/_PP-iPPPPPPPPPPPPPPPPPP-iP-i0[-i-iP-iP1VPP2UPP2UPP3P$OP3W3^3d3j3p3z4Q4WPPPP4^4^4jP4mP4u4x5Q5TTQORTPORQ[SQe]V!`m!b#lRcYQldQ#m!dQ$s#jR%l$tSPORQaVT#O{#QQ!XgQ#_!TQ#e!WQ#h!]Q#i!^Q$O!tQ$p#gQ%Z$fQ%c$kR&e&^QqfQ!fpS!xw%^Q!zzQ!{{Q#[!QQ#]!RQ#q!gQ#r!hQ#s!iQ#t!jQ#u!kQ#v!lQ#w!mQ#x!nQ#y!oQ#z!pQ#{!qQ#|!rQ$P!yQ$V#QQ$X#SQ$^#WQ$o#fY$v#n$x${%n%o[%P$h%x%z&X&[&dQ%q%TQ%r%VR&g&h!h}fpwz!Q!R!g!h!i!j!k!l!m!n!o!p!q!r!y#S#W#f#n$h$x${%T%V%^%n%o%z&[&d&hS!}{#QQ&W%xR&b&X!p|fpwz{!Q!R!g!h!i!j!k!l!m!n!o!p!q!r!y#Q#S#W#f#n$h$x${%T%V%^%n%o%x%z&X&[&d&hQ![iQ#^!SQ$]#VQ$e#ZQ$g#[Q$q#hQ%S$^Q%U&gQ%v%ZS%w%[&iQ%}%bR&T%qT#P{#Q!h}fpwz!Q!R!g!h!i!j!k!l!m!n!o!p!q!r!y#S#W#f#n$h$x${%T%V%^%n%o%z&[&d&hS!|{#QQ&V%xR&a&X!fyfpwz{!Q!R!g!h!i!j!k!l!m!n!o!p!q!r!y#Q#S#W#f#n$x${%T%V%n%o%x&X&hW%_$h%z&[&dR%y%^R!wr!qsfpwz{!Q!R!g!h!i!j!k!l!m!n!o!p!q!r!y#Q#S#W#f#n$h$x${%T%V%^%n%o%x%z&X&[&d&hQi_R#Z!P!qyfpwz{!Q!R!g!h!i!j!k!l!m!n!o!p!q!r!y#Q#S#W#f#n$h$x${%T%V%^%n%o%x%z&X&[&d&h!qufpwz{!Q!R!g!h!i!j!k!l!m!n!o!p!q!r!y#Q#S#W#f#n$h$x${%T%V%^%n%o%x%z&X&[&d&h!hwfpwz{!Q!R!g!h!i!j!k!l!m!n!o!p!q!r!y#Q#S#W#f#n$x${%T%V%^%n%o%x&X&hX%^$h%z&[&d!qxfpwz{!Q!R!g!h!i!j!k!l!m!n!o!p!q!r!y#Q#S#W#f#n$h$x${%T%V%^%n%o%x%z&X&[&d&h!p|fpwz{!Q!R!g!h!i!j!k!l!m!n!o!p!q!r!y#Q#S#W#f#n$h$x${%T%V%^%n%o%x%z&X&[&d&hT%w%[&i!q|fpwz{!Q!R!g!h!i!j!k!l!m!n!o!p!q!r!y#Q#S#W#f#n$h$x${%T%V%^%n%o%x%z&X&[&d&hSPORRaVQRORWRQneR!cnQ$m#bR%f$mQ#Q{R$W#QQ#o!fS$y#o$|R$|$PQ%X$bR%u%XQ%i$pR&Q%iQ%{%aR&]%{e!Vg!T!W!]!^!t#g$f$k&^R#d!UQ#b!UV%d$l%e&OR$d#YQ$b#YV%s%W%t&UR$R!yQ%a$hV&Y%z&[&d",
  nodeNames: "⚠ LineComment BlockComment Program Item ImportItem Ident import ImportItemCanditate type TypeImport templ trigger { , } from Module ; LetStatement let Type Word [ ] String QuestionMark Expression ExpressionWithBlock Block Statement TypeDefinition ReturnStatement return BreakStatement break ContinueStatement continue LoopStatement loop WhileStatement while ForStatement for in ExpressionStatement ExpressionWithoutBlock LiteralExpression Number Boolean Boolean ListLiteral ObjectLiteral new FunctionLiteral fn Parameters ( ) Arrow null none RangeExpression Range VariableName BuiltinFunc BuiltinFunc PrefixExpression PrefixOp Minus Not InfixExpression BitOr BitXor BitAnd ShiftLeft ShiftRight Plus Multiply Divide Modulo Power LogicalAnd LogicalOr Equal NotEqual LessThan LessThanEqual GreaterThan GreaterThanEqual AssignExpression AssignOp CallExpression CallBase IndexExpression MemberExpression . CastExpression as IfExpression if else MatchExpression match FatArrow TryExpression try catch FunctionDefinition PubItem pub",
  maxTerm: 132,
  skippedNodes: [0,1,2],
  repeatNodeCount: 8,
  tokenData: "/P~RqXY#YYZ#Y]^#Ypq#Yqr#_rs#luv%^vw%kwx%{xy'hyz'mz{'r{|(V|}(_}!O(d!O!P(t!P!Q)R!Q![+P![!]+r!]!^+w!^!_+|!_!`,f!`!a,}!a!b-g!c!}-n!}#O.R#P#Q.W#Q#R.]#R#S-n#T#o-n#o#p.e#p#q.j#q#r.z~#_O#m~~#dP!h~!_!`#g~#lO!w~~#qVi~Or#lrs$Ws#O#l#O#P$]#P;'S#l;'S;=`%W<%lO#l~$]Oi~~$`RO;'S#l;'S;=`$i;=`O#l~$nWi~Or#lrs$Ws#O#l#O#P$]#P;'S#l;'S;=`%W;=`<%l#l<%lO#l~%ZP;=`<%l#l~%cP!r~!_!`%f[%kO!}[~%pQ!l~vw%v!_!`%f~%{O!t~~&QVi~Ow%{wx$Wx#O%{#O#P&g#P;'S%{;'S;=`'b<%lO%{~&jRO;'S%{;'S;=`&s;=`O%{~&xWi~Ow%{wx$Wx#O%{#O#P&g#P;'S%{;'S;=`'b;=`<%l%{<%lO%{~'eP;=`<%l%{~'mO!Z~~'rO![~~'wQ!p~z{'}!_!`%f~(SP!s~!_!`%f~([P!o~!_!`%f~(dO^~_(iQ!g]!_!`%f!`!a(oQ(tO!]Q~(yP#S~!O!P(|~)RO!a~~)WR!q~z{)a!P!Q*h!_!`%f~)dTOz)az{)s{;'S)a;'S;=`*b<%lO)a~)vVOz)az{)s{!P)a!P!Q*]!Q;'S)a;'S;=`*b<%lO)a~*bOQ~~*eP;=`<%l)a~*mSP~OY*hZ;'S*h;'S;=`*y<%lO*h~*|P;=`<%l*h~+UR!Q~!O!P+_!Q![+P#Y#Z+m~+bP!Q![+e~+jP!Q~!Q![+e~+rO!Q~~+wO#n~~+|Ob~~,RQ!x~!^!_,X!_!`,a~,^P!m~!_!`%f~,fO!y~],mQ#tP!}[!_!`,s!`!a,x[,xO!v[S,}O#[S~-SQ!z~!_!`-Y!`!a-_~-_O!{~~-dP!n~!_!`%f_-nO#rSjZ_-uSU]fQ!Q![-n!c!}-n#R#S-n#T#o-n~.WOg~~.]Oh~~.bP!k~!_!`%f~.jO]~~.oQ!j~!_!`%f#p#q.u~.zO!u~~/PO_~",
  tokenizers: [0, 1, 2, 3],
  topRules: {"Program":[0,3]},
  specialized: [{term: 6, get: (value) => spec_Ident[value] || -1}],
  tokenPrec: 2656
});

const HomescriptLanguage = language.LRLanguage.define({
    parser: parser.configure({
        props: [
            language.indentNodeProp.add({
                Application: language.delimitedIndent({ closing: ')', align: false }),
            }),
            language.foldNodeProp.add({
                Application: language.foldInside,
            }),
            highlight.styleTags({
                'for while loop if else match try catch return break continue': highlight.tags.controlKeyword,
                'in new': highlight.tags.operatorKeyword,
                'let fn type templ trigger': highlight.tags.definitionKeyword,
                'pub': highlight.tags.modifier,
                'import from': highlight.tags.moduleKeyword,
                'ImportItem/Module/Ident': highlight.tags.namespace,
                // 'ImportItem/ImportItemCanditate/Ident': t.namespace,
                'ImportItem/ImportItemCanditate/TypeImport/Ident': highlight.tags.typeName,
                'TypeDefinition/Ident': highlight.tags.namespace,
                'FunctionDefinition/Ident': highlight.tags.function(highlight.tags.variableName),
                as: highlight.tags.keyword,
                Boolean: highlight.tags.bool,
                null: highlight.tags.null,
                none: highlight.tags.null,
                'CallExpression/CallBase/Expression/ExpressionWithoutBlock/VariableName/Ident': highlight.tags.function(highlight.tags.propertyName),
                'ObjectLiteral/Ident': highlight.tags.propertyName,
                'MemberExpression/Ident': highlight.tags.propertyName,
                'FunctionDefinition/parameterList/Ident': highlight.tags.local(highlight.tags.variableName),
                'ForStatement/Ident': highlight.tags.local(highlight.tags.variableName),
                'Parameters/Ident': highlight.tags.local(highlight.tags.variableName),
                'VariableName/Ident': highlight.tags.variableName,
                LineComment: highlight.tags.lineComment,
                BlockComment: highlight.tags.blockComment,
                Number: highlight.tags.number,
                String: highlight.tags.string,
                'Arrow QuestionMark': highlight.tags.typeOperator,
                'FatArrow': highlight.tags.controlOperator,
                'Plus Minus Multiply Divide Modulo Power': highlight.tags.arithmeticOperator,
                'LogicalOr LogicalAnd': highlight.tags.logicOperator,
                'LessThan LessThanEqual GreaterThan GreaterThanEqual NotEqual Equal': highlight.tags.compareOperator,
                'AssignOp`': highlight.tags.definitionOperator,
                '"(" ")" "{" "}" "[" "]"': highlight.tags.bracket,
                '"." ".." "," ";"': highlight.tags.separator,
                BuiltinFunc: highlight.tags.standard(highlight.tags.function(highlight.tags.variableName)),
                'Type/Word': highlight.tags.typeName,
            }),
        ],
    }),
    languageData: {
        commentTokens: { line: '//' },
    },
});
const HomescriptCompletion = HomescriptLanguage.data.of({
    autocomplete: autocomplete.completeFromList([
        { label: 'pub', type: 'keyword' },
        { label: 'new', type: 'keyword' },
        { label: 'fn', type: 'keyword' },
        { label: 'let', type: 'keyword' },
        { label: 'return', type: 'keyword' },
        { label: 'break', type: 'keyword' },
        { label: 'continue', type: 'keyword' },
        { label: 'if', type: 'keyword' },
        { label: 'else', type: 'keyword' },
        { label: 'match', type: 'keyword' },
        { label: 'loop', type: 'keyword' },
        { label: 'while', type: 'keyword' },
        { label: 'for', type: 'keyword' },
    ]),
});
function Homescript() {
    return new language.LanguageSupport(HomescriptLanguage, [HomescriptCompletion]);
}

exports.Homescript = Homescript;
exports.HomescriptCompletion = HomescriptCompletion;
exports.HomescriptLanguage = HomescriptLanguage;
