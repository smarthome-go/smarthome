'use strict';

Object.defineProperty(exports, '__esModule', { value: true });

var language = require('@codemirror/language');
var autocomplete = require('@codemirror/autocomplete');
var highlight = require('@lezer/highlight');
var lr = require('@lezer/lr');

// This file was generated by lezer-generator. You probably shouldn't edit it.
const spec_Ident = {__proto__:null,import:14, type:18, templ:22, trigger:26, from:36, let:44, return:74, break:78, continue:82, loop:86, while:90, for:94, in:96, true:108, false:108, on:108, off:108, new:114, fn:118, null:128, none:130, exit:140, throw:140, assert:140, print:140, println:140, debug:140, as:204, if:208, else:210, match:214, try:220, catch:222, pub:228, impl:232, event:240};
const parser = lr.LRParser.deserialize({
  version: 14,
  states: "HbQ]QPOOOzQPO'#EvO!PQPO'#EwO!UQPO'#EyOOQO'#C`'#C`OOQO'#E|'#E|Q]QPOOO!ZQPO'#CaO!lQPO'#DPO!qQPO'#CqO!vQPO'#EpO!{QPO'#EqO#WQPO'#EsO#RQPO'#C`OOQO,5;b,5;bO#]QPO,5;cO#bQPO,5;eOOQO-E8z-E8zOOQO'#Cd'#CdO#mQPO'#CdO#rQPO'#CdO#wQPO'#CdO#|QPO,58{O$RQPO,58{O$aQPO,59kO$fQPO,59]O$nQPO,5;[OOQO,5;],5;]OOQO'#Eu'#EuO$sQPO,5;_OOQO,58z,58zO$xQPO'#F]O%ZQPO1G0}O%`QPO1G1PO%hQPO1G1POOQO1G1P1G1PO%hQPO1G1PO%pQQO'#EzOOQO'#Cf'#CfOOQO,59O,59OOOQO'#Ch'#ChOOQO'#Cj'#CjO%{QPO1G.gO&QQPO1G.gO&YQSO1G/VO&hQWO1G.wO&YQSO1G.wO'oQPO'#DjO'wQPO1G0vO(PQPO1G0yO(UQPO'#CwOOQO'#Cx'#CxO(ZQPO'#F`O(`QPO'#F`OOQO'#F^'#F^O(hQPO'#F^OOQO,5;w,5;wO(pQPO,5;wOOQO7+&i7+&iOOQO,5;q,5;qOOQO7+&k7+&kO(uQPO7+&kOOQO-E9T-E9TO(}QPO7+&kOOQO'#E{'#E{O&hQWO,5;fOOQO'#Co'#CoO)VQPO7+$RO)[QPO7+$RO)mQPO7+$RO)uQPO7+$RO&YQSO'#F[OOQO'#Cs'#CsO&YQSO'#CsO)zQPO7+$qO*PQWO'#DeO+wQWO7+$cO,OQPO'#DaOOQO'#Da'#DaOOQO'#Ds'#DsOOQO'#Dr'#DrOOQO'#Dv'#DvO&hQWO'#DuO,TQPO'#E`OOQO'#D`'#D`O&hQWO'#D`O,YQWO'#C}OOQO'#C|'#C|OOQO'#C{'#C{O,yQPO'#DfO$nQPO'#DhO&hQWO'#EgO&hQWO'#EjO-OQPO'#EmO-TQPO7+$cOOQO,5:U,5:UO-YQPO,5:UOOQO7+&b7+&bO&YQSO7+&bO-_QPO7+&eOOQO,59c,59cO&YQSO,5;zO-dQPO,5;zO-iQPO,5;xO-wQPO,5;xOOQO1G1c1G1cOOQO<<JV<<JVP.PQPO'#FVO.UQPO<<JVO*WQWOOO/}QWO1G1QOOQO<<Gm<<GmOOQO,5;i,5;iO0XQPO<<GmO0^QPO<<GmOOQO-E8{-E8{O%{QPO<<GmO0oQPO,5;vOOQO,59_,59_OOQO<<H]<<H]OOQO,5:P,5:PO0tQWO,5:PO&hQWO,5:[O&hQWO,5:eO&hQWO,5:eO&hQWO,5:eO&hQWO,5:eO&hQWO,5:eO&hQWO,5:eO&hQWO,5:eO&hQWO,5:eO&hQWO,5:eO&hQWO,5:xO&hQWO,5:|O1OQPO,5:}O&YQSO,5;POOQO<<G}<<G}OOQO'#Dc'#DcOOQO,59{,59{O1TQWO,5:aO3ZQWO,5:zO3bQWO,59zO3iQWO,59iO5aQWO'#C{O7RQWO'#D_OOQO'#DO'#DOOOQO'#FP'#FPO8zQWO,59iOOQO,59i,59iO9kQWO'#DQO9rQPO'#DSO9wQPO'#DUO-OQPO'#DWO&hQWO'#DYO9|QPO'#D[O:RQPO,5:QO:aQPO,5:SO:iQWO,5;RO:pQWO,5;UO:wQPO,5;XO&hQWO<<G}O&YQSO1G/pO-OQPO<<I|O:|QPO<<JPOOQO1G1f1G1fO&YQSO1G1fOOQO,5;j,5;jO;XQPO1G1dOOQO-E8|-E8|OOQOAN?qAN?qO%{QPOAN=XO;gQPOAN=XP$RQPO'#E}O;lQPOAN=XOOQO1G1b1G1bO;qQWO1G/kO;xQPO1G/kOOQO1G/k1G/kO<QQWO1G/vO@ZQWO1G0PO@bQWO1G0POBiQWO1G0POBsQWO1G0PODzQWO1G0POEXQWO1G0POGiQWO1G0POGpQWO1G0POHxQWO1G0POJRQWO1G0dOK[QWO1G0hOOQO1G0i1G0iOOQO1G0k1G0kOKcQWO'#FeOOQO1G0f1G0fOKmQPO1G0fOOQO1G/f1G/fOOQO1G/T1G/TOOQO,59y,59yOKrQWO1G/TOOQO-E8}-E8}OKyQWO,59lOOQO,59l,59lOOQO,59n,59nOOQO,59p,59pOLQQWO,59rO:iQWO,59tOMwQPO,59vOM|QPO'#FdOOQO'#Fc'#FcONRQPO'#FcOOQO1G/l1G/lONZQPO1G/lOOQO1G/n1G/nO&YQSO1G/nON`QWO1G0mO!%RQWO1G0pO!%YQPO1G0sO!%_QWOAN=iO!%fQPO7+%[OOQOAN?hAN?hOOQO'#FU'#FUO!%nQPOAN?kO#RQPO'#FUOOQOAN?kAN?kOOQO7+'Q7+'QP!%yQPO'#FOO!&UQPOG22sO%{QPOG22sOOQOG22sG22sO!&ZQWO,5;lOOQO7+%V7+%VO!&hQWO7+%VOOQO-E9O-E9OOOQO7+&S7+&SO!&oQWO,5<PO!&vQPO,5<POOQO7+&Q7+&QOOQO7+$o7+$oOOQO1G/W1G/WOOQO1G/^1G/^O!'OQWO1G/`O&hQWO1G/bO&hQWO,5<OO!(uQPO,5;}O!)QQPO,5;}OOQO7+%W7+%WO-OQPO7+%YO!)YQPO7+&XO!)bQWO'#DrO&hQWO'#DuO!+UQWO'#D`OOQO7+&[7+&[O!+]QPO7+&[O-OQPO7+&_OOQOG23TG23TO!+eQPO'#FSO!+jQPO<<HvOOQO<<Hv<<HvOOQO-E9S-E9SOOQOG25VG25VOOQO,5;p,5;pOOQOLD(_LD(_O!+rQPOLD(_OOQO<<Hq<<HqP&hQWO'#FQO!+wQWO1G1kOOQO7+$z7+$zO:iQWO7+$|O!,OQWO1G1jOOQO,5;m,5;mO!,YQPO1G1iOOQO-E9P-E9POOQO<<Ht<<HtOOQO<<Is<<IsO&hQWO,5<QO!,eQWO,5<QO!,lQWO<<IvO!,sQPO<<IvOOQO<<Iv<<IvOOQO<<Iy<<IyO!,{QPO,5;nOOQO-E9Q-E9QOOQOAN>bAN>bOOQO!$'Ky!$'KyO!-QQWO<<HhP!.wQPO'#FRO!/PQWO1G1lO!/WQWO1G1lO&hQWO1G1lOOQO,5;o,5;oOOQOAN?bAN?bO!/bQWOAN?bOOQO-E9R-E9RO&YQSO1G1YOOQOAN>SAN>SOOQO7+'W7+'WO!/iQWO7+'WO!/pQWO7+'WOOQOG24|G24|P!#zQWO'#FTOOQO7+&t7+&tOOQO<<Jr<<Jr",
  stateData: "!/}~O#|OSPOSQOS~OVVOXWOfXO!]YO#fZO#h[O#l]O$ZPO$[RO~OU^O~O$U_O~Oi`O~OUbOXcOZdO]eO_gO~OUhO~OUiO~OUjO~OXWOfXO!]YO~OUlO~O_oO~O]uO`qOjsO~OUvO~OUxO~OUyO~ObzO~OUbOXcOZdO]eO~O$U|O~O#}!OO$U}O~O!_!PO~O!P!RO~OU!TOa!YOm!TO$R!WO$T!SO~Od![O~O]uOj!^O~O`!_Oj!^O~O$]!bO$^!bO$_!bO~OU!dO~O`!fOa!hO~O_oOh!jOi!iOn!kO~OU!rO_!xOi!mOm!pOn!sO!U!oO!Z!{O!]!|O!_!wO!b!pO!c!pO!h!qO!k!sO!l!sO#[!}O#_#OO#b#PO~OU#SO!`#RO~O_!xO!a#UO~O$ZPO~OU#WO~O#}#XO~OU!TOm!TO~O`#ZOa$QX~Oa#]O~O]uOj#^O~O`#`Oj#^O~Od#cO~OUbOXcOZdO]eOa#eO~O`#fOa#eO~Ob#hO~Od#kO~Oj#lO~P&hOi#yO!e#nO!k#sO!n#oO!o#pO!p#qO!q#rO!r#rO!s#sO!t#tO!u#tO!v#tO!w#uO!x#vO!y#uO!z#wO!{#wO!|#wO!}#wO#O#wO#P#wO#R#xO#W#zO#Y#{O!_#TX~Od#|O~P*WO!W#}O~O!_$QO~OXWOa$YOfXOu$ZOw$[Oy$]O{$^O}$_O!P$`O~P&hO_$aO~O_!xO~O$U$fO~O#}$gO~O_$iO~O#}$kO~OU!TOm!TO$T!SOa$Qa~O`$mOa$Qa~O]uO~O]uOj$oO~Oi!SX!_!SX!e!SX!k!SX!n!SX!o!SX!p!SX!q!SX!r!SX!s!SX!t!SX!u!SX!v!SX!w!SX!x!SX!y!SX!z!SX!{!SX!|!SX!}!SX#O!SX#P!SX#R!SX#W!SX#Y!SX~O`#nij#ni~P.^Ob$pO~OUbOXcOZdO]eOa$qO~Oj$tO~O`$uOj$wO~P*WOU%UO~O!w#uO#W#zOd!iai!ia!_!ia!e!ia!k!ia!n!ia!o!ia!p!ia!q!ia!r!ia!s!ia!t!ia!u!ia!v!ia!x!ia!y!ia!z!ia!{!ia!|!ia!}!ia#O!ia#P!ia#R!ia#Y!ia`!iaj!ia!`!iaa!ia_!ia#`!ia~O!`%XO~P&hO!`%ZO~P*WOa%[O~P*WOioX!_oX!eoX!koX!noX!ooX!poX!qoX!roX!soX!toX!uoX!voX!woX!xoX!yoX!zoX!{oX!|oX!}oX#OoX#PoX#RoX#WoX#YoX~Od%]OaoX~P3pO!eoX!noX!ooX!poX!qoX!roX!soX!toX!uoX!voX!woX!xoX!yoX!zoX!{oX!|oX!}oX#OoX#PoX#RoX#WoX#YoX~Od%]OU!RXX!RX_!RXa!RXf!RXi!RXm!RXn!RXu!RXw!RXy!RX{!RX}!RX!P!RX!U!RX!Z!RX!]!RX!_!RX!b!RX!c!RX!h!RX!k!RX!l!RX#[!RX#_!RX#b!RX~P5kOXWOa%[OfXOu$ZOw$[Oy$]O{$^O}$_O!P$`O~P&hOd%aO~P&hOd%bO~Od%cO~OU%fO~OU%gOa%jOm%gO$R%hO~O_!xO!a%mO~O_!xO~P*WO_%oO~P*WO#c%pO~Oa%wO!]YO#f%vO~OU!TOm!TO$T!SOa$Qi~Ob%{O~Od%|O~Oj&OO~P&hO`&POj&OO~O!w#uO!x#vO!y#uO#W#zOd!dii!di!_!di!e!di!k!di!n!di!o!di!p!di!q!di!r!di!s!di!t!di!u!di!v!di!z!di!{!di!|!di!}!di#O!di#P!di#R!di#Y!di`!dij!di!`!dia!di_!di#`!di~O!e#nO!k#sO!p#qO!q#rO!r#rO!s#sO!t#tO!u#tO!v#tO!w#uO!x#vO!y#uO#W#zOd!mii!mi!_!mi!n!mi!z!mi!{!mi!|!mi!}!mi#O!mi#P!mi#R!mi#Y!mi`!mij!mi!`!mia!mi_!mi#`!mi~O!o#pO~P>WO!o!mi~P>WO!e#nO!k#sO!s#sO!t#tO!u#tO!v#tO!w#uO!x#vO!y#uO#W#zOd!mii!mi!_!mi!n!mi!o!mi!p!mi!z!mi!{!mi!|!mi!}!mi#O!mi#P!mi#R!mi#Y!mi`!mij!mi!`!mia!mi_!mi#`!mi~O!q#rO!r#rO~P@iO!q!mi!r!mi~P@iO!e#nO!w#uO!x#vO!y#uO#W#zOd!mii!mi!_!mi!k!mi!n!mi!o!mi!p!mi!q!mi!r!mi!s!mi!z!mi!{!mi!|!mi!}!mi#O!mi#P!mi#R!mi#Y!mi`!mij!mi!`!mia!mi_!mi#`!mi~O!t#tO!u#tO!v#tO~PB}O!t!mi!u!mi!v!mi~PB}O!w#uO#W#zOd!mii!mi!_!mi!e!mi!k!mi!n!mi!o!mi!p!mi!q!mi!r!mi!s!mi!t!mi!u!mi!v!mi!x!mi!z!mi!{!mi!|!mi!}!mi#O!mi#P!mi#R!mi#Y!mi`!mij!mi!`!mia!mi_!mi#`!mi~O!y!mi~PEfO!y#uO~PEfO!e#nO!k#sO!n#oO!o#pO!p#qO!q#rO!r#rO!s#sO!t#tO!u#tO!v#tO!w#uO!x#vO!y#uO#W#zO~Od!mii!mi!_!mi!z!mi!{!mi!|!mi!}!mi#O!mi#P!mi#R!mi#Y!mi`!mij!mi!`!mia!mi_!mi#`!mi~PGwO!z#wO!{#wO!|#wO!}#wO#O#wO#P#wO#R#xOd#Qii#Qi!_#Qi#Y#Qi`#Qij#Qi!`#Qia#Qi_#Qi#`#Qi~PGwOj&RO~P*WO`&SO!`$XX~P*WO!`&UO~Oa&VO~P*WOd&WO~P*WOd&XOUzaXza_zaazafzaizamzanzauzawzayza{za}za!Pza!Uza!Zza!]za!_za!bza!cza!hza!kza!lza#[za#_za#bza~O!Q&ZO~O#}&[O~O`&]Oa$VX~Oa&_O~O#]&aOd#Zii#Zi!_#Zi!e#Zi!k#Zi!n#Zi!o#Zi!p#Zi!q#Zi!r#Zi!s#Zi!t#Zi!u#Zi!v#Zi!w#Zi!x#Zi!y#Zi!z#Zi!{#Zi!|#Zi!}#Zi#O#Zi#P#Zi#R#Zi#W#Zi#Y#Zi`#Zij#Zi!`#ZiU#ZiX#Zi_#Zia#Zif#Zim#Zin#Ziu#Ziw#Ziy#Zi{#Zi}#Zi!P#Zi!U#Zi!Z#Zi!]#Zi!b#Zi!c#Zi!h#Zi!l#Zi#[#Zi#_#Zi#b#Zi#`#Zi~OU&bO_!xOi!mOm!pOn!sO!U!oO!Z!{O!]!|O!_!wO!b!pO!c!pO!h!qO!k!sO!l!sO#[!}O#_#OO#b#PO~Oa&eO~P!#zOU&gO~Od&hO~P*WO`&iO!`&kO~Oa&mO!]YO#f%vO~OU!TOm!TO$T!SO~Od&oO~O`#taj#ta!`#ta~P*WOj&qO~P&hO!`$Xa~P&hO`&sO!`$Xa~Od&tOU|iX|i_|ia|if|ii|im|in|iu|iw|iy|i{|i}|i!P|i!U|i!Z|i!]|i!_|i!b|i!c|i!h|i!k|i!l|i#[|i#_|i#b|i~OU%gOm%gOa$Va~O`&xOa$Va~O_!xO#[!}O~O#`&|Oi!fX!_!fX!e!fX!k!fX!n!fX!o!fX!p!fX!q!fX!r!fX!s!fX!t!fX!u!fX!v!fX!w!fX!x!fX!y!fX!z!fX!{!fX!|!fX!}!fX#O!fX#P!fX#R!fX#W!fX#Y!fX~O#`&|O~P.^O`'OOa'QO~OU'SO~O`&iO!`'UO~Od'VO~O!`$Xi~P&hO`$Wia$Wi~P*WOU%gOm%gOa$Vi~O#`'[O~P.^Oa'^O~P!#zO`'_Oa'^O~O#}'aO~Od'bOU!OyX!Oy_!Oya!Oyf!Oyi!Oym!Oyn!Oyu!Oyw!Oyy!Oy{!Oy}!Oy!P!Oy!U!Oy!Z!Oy!]!Oy!_!Oy!b!Oy!c!Oy!h!Oy!k!Oy!l!Oy#[!Oy#_!Oy#b!Oy~OU%gOm%gO~O`'cO~P3pO`$Yia$Yi~P3pOa'fO~P!#zO`'iO~P3pO`$Yqa$Yq~P3pOP!u~",
  goto: "8`$ZPPPP$[$`PP$dP$oP$rP$uPPPP$xP%UP%aPPP&S&YPP&c(f)g+T%U%]P%]P%]P%]P%]P%]PP%]+X,YP-XP-[-[P-[P.VPPPPP-[P.]/WP.]0RPP.]PPPPPPPPPPPPPPPPPP.]P1O1{.].]P.]P2vPP3tPP3tPP4o$`P$`P5Q5T$`P$`5[5e5h5n5t5z6Q6[6b6h6n6tPPPP7O7]7mP7pPP7x7{8T8WTTOUTSOUQfVQ{gV#d!f#f$rRwcRwdRweQ!ezQ$s#hQ%z$pR&p%{SSOUQkZT$V!x$XQ!l|Q#Q!OQ#i!iQ#j!kQ$h#UQ$j#XQ%V#{Q%r$gQ%x$kQ&`%mR'h'aX!Vo#Z$m%yW!Uo#Z$m%yR#Y!VQ!n}^#a!c%o&|'O'['_'gQ#m!mS$P!t&cQ$R!wQ$S!xQ$c!}Q$d#OQ$x#nQ$y#oQ$z#pQ${#qQ$|#rQ$}#sQ%O#tQ%P#uQ%Q#vQ%R#wQ%S#xQ%T#yQ%W$QQ%^$XQ%`$ZQ%e$_Q%q$fY%}$u&P&S&r&sQ&u&ZR&v&[!h!z}!c!m!t!w!}#O#n#o#p#q#r#s#t#u#v#w#x#y$Q$Z$_$f$u%o&P&S&Z&[&c&r&s'O'_'gS$U!x$XQ'Z&|R'e'[!p!y}!c!m!t!w!x!}#O#n#o#p#q#r#s#t#u#v#w#x#y$Q$X$Z$_$f$u%o&P&S&Z&[&c&r&s&|'O'['_'gQ#T!QQ$e#PQ%d$^Q%l$bQ%n$cQ%s$hQ&Y%eQ&z&`Q&{&aQ'R&gR'W&uT$W!x$X!h!z}!c!m!t!w!}#O#n#o#p#q#r#s#t#u#v#w#x#y$Q$Z$_$f$u%o&P&S&Z&[&c&r&s'O'_'gS$T!x$XQ'Y&|R'd'[!f!v}!c!m!t!w!x!}#O#n#o#p#q#r#s#t#u#v#w#x#y$Q$X$Z$_$f$u&P&S&Z&[&r&s&|'[W&d%o'O'_'gR&}&cR$O!o!q!p}!c!m!t!w!x!}#O#n#o#p#q#r#s#t#u#v#w#x#y$Q$X$Z$_$f$u%o&P&S&Z&[&c&r&s&|'O'['_'gQ!QjR$b!|!q!v}!c!m!t!w!x!}#O#n#o#p#q#r#s#t#u#v#w#x#y$Q$X$Z$_$f$u%o&P&S&Z&[&c&r&s&|'O'['_'g!q!r}!c!m!t!w!x!}#O#n#o#p#q#r#s#t#u#v#w#x#y$Q$X$Z$_$f$u%o&P&S&Z&[&c&r&s&|'O'['_'g!h!t}!c!m!t!w!x!}#O#n#o#p#q#r#s#t#u#v#w#x#y$Q$X$Z$_$f$u&P&S&Z&[&c&r&s&|'[X&c%o'O'_'g!n!v}!m!t!w!x!}#O#n#o#p#q#r#s#t#u#v#w#x#y$Q$X$Z$_$f$u%o&P&S&Z&[&c&r&s&|'O'['_'gR#b!c!q!u}!c!m!t!w!x!}#O#n#o#p#q#r#s#t#u#v#w#x#y$Q$X$Z$_$f$u%o&P&S&Z&[&c&r&s&|'O'['_'g!p!y}!c!m!t!w!x!}#O#n#o#p#q#r#s#t#u#v#w#x#y$Q$X$Z$_$f$u%o&P&S&Z&[&c&r&s&|'O'['_'gR&{&a!q!y}!c!m!t!w!x!}#O#n#o#p#q#r#s#t#u#v#w#x#y$Q$X$Z$_$f$u%o&P&S&Z&[&c&r&s&|'O'['_'gSSOUQkZQn]S%t$i%uR&n%vRm[SQOUR#V!RQt`X!]q!_#_#`R!cuQUORaUQ!g{R#g!gQ#[!XR$n#[Q$X!xR%_$XQ$v#mS&Q$v&TR&T%WQ&^%iR&y&^Q&j%rR'T&jQ'P&fR'`'PQ%u$iR&l%uQr`S!`r!aR!atg!j|!O!i!k#U#X#{$g$k%m'aQp_g!j|!O!i!k#U#X#{$g$k%m'aR!ZoQ!XoV$l#Z$m%yR%k$aQ%i$aV&w&]&x'XR%Y$QQ&f%oV']'O'_'g",
  nodeNames: "⚠ LineComment BlockComment Program Item ImportItem Ident import ImportItemCanditate type TypeImport templ TemplImport trigger TriggerImport { , } from Module ; LetStatement let Type Word [ ] ObjectTypeFieldAnnotation ObjectTypeFieldKey String QuestionMark Expression ExpressionWithBlock Block Statement TypeDefinition ReturnStatement return BreakStatement break ContinueStatement continue LoopStatement loop WhileStatement while ForStatement for in ExpressionStatement ExpressionWithoutBlock LiteralExpression Number Boolean Boolean ListLiteral ObjectLiteral new FunctionLiteral fn Parameters ( ) Arrow null none RangeExpression Range VariableName BuiltinFunc BuiltinFunc PrefixExpression PrefixOp Minus Not InfixExpression BitOr BitXor BitAnd ShiftLeft ShiftRight Plus Multiply Divide Modulo Power LogicalAnd LogicalOr Equal NotEqual LessThan LessThanEqual GreaterThan GreaterThanEqual AssignExpression AssignOp CallExpression CallBase IndexExpression MemberExpression . CastExpression as IfExpression if else MatchExpression match FatArrow TryExpression try catch FunctionDefinition PubItem pub ImplBlock impl ImplTemplateIdent SingletonIdent SingletonDefinition event Annotation TriggerAnnotationItem TriggerConnective",
  maxTerm: 153,
  skippedNodes: [0,1,2],
  repeatNodeCount: 10,
  tokenData: "2x~RyXY#rYZ#r]^#rpq#rqr#wrs$Ust%vtu%{uv&Qvw&_wx&oxy([yz(az{(f{|(y|})R}!O)W!O!P)h!P!Q)u!Q![+s![!],f!]!^,k!^!_,p!_!`-Y!`!a-q!a!b.Z!b!c.b!c!}.g!}#O.z#P#Q/P#Q#R/U#R#S.g#T#U/^#U#].g#]#^0^#^#c.g#c#d1^#d#o.g#o#p2^#p#q2c#q#r2s~#wO#|~~#|P!l~!_!`$P~$UO!{~~$ZVm~Or$Urs$ps#O$U#O#P$u#P;'S$U;'S;=`%p<%lO$U~$uOm~~$xRO;'S$U;'S;=`%R;=`O$U~%WWm~Or$Urs$ps#O$U#O#P$u#P;'S$U;'S;=`%p;=`<%l$U<%lO$U~%sP;=`<%l$U~%{O$[~~&QO$Z~~&VP!v~!_!`&YW&_O#RW~&dQ!p~vw&j!_!`&Y~&oO!x~~&tVm~Ow&owx$px#O&o#O#P'Z#P;'S&o;'S;=`(U<%lO&o~'^RO;'S&o;'S;=`'g;=`O&o~'lWm~Ow&owx$px#O&o#O#P'Z#P;'S&o;'S;=`(U;=`<%l&o<%lO&o~(XP;=`<%l&o~(aO!_~~(fO!`~~(kQ!t~z{(q!_!`&Y~(vP!w~!_!`&Y~)OP!s~!_!`&Y~)WO`~X)]Q!kW!_!`&Y!`!a)cP)hO!aP~)mP#W~!O!P)p~)uO!e~~)zR!u~z{*T!P!Q+[!_!`&Y~*WTOz*Tz{*g{;'S*T;'S;=`+U<%lO*T~*jVOz*Tz{*g{!P*T!P!Q+P!Q;'S*T;'S;=`+U<%lO*T~+UOQ~~+XP;=`<%l*T~+aSP~OY+[Z;'S+[;'S;=`+m<%lO+[~+pP;=`<%l+[~+xR!U~!O!P,R!Q![+s#Y#Z,a~,UP!Q![,X~,^P!U~!Q![,X~,fO!U~~,kO#}~~,pOd~~,uQ!|~!^!_,{!_!`-T~-QP!q~!_!`&Y~-YO!}~X-aQ$UP#RW!_!`-g!`!a-lW-lO!zWW-qO#`W~-vQ#O~!_!`-|!`!a.R~.RO#P~~.WP!r~!_!`&Y].bO$RPn[~.gO$T~].nSUXhS!Q![.g!c!}.g#R#S.g#T#o.g~/POi~~/UOj~~/ZP!o~!_!`&Y_/eUUXhS!Q![.g!c!}.g#R#S.g#T#h.g#h#i/w#i#o.g_0QS$^QUXhS!Q![.g!c!}.g#R#S.g#T#o.g_0eUUXhS!Q![.g!c!}.g#R#S.g#T#b.g#b#c0w#c#o.g_1QS$_QUXhS!Q![.g!c!}.g#R#S.g#T#o.g_1eUUXhS!Q![.g!c!}.g#R#S.g#T#b.g#b#c1w#c#o.g_2QS$]QUXhS!Q![.g!c!}.g#R#S.g#T#o.g~2cO_~~2hQ!n~!_!`&Y#p#q2n~2sO!y~~2xOa~",
  tokenizers: [0, 1, 2, 3],
  topRules: {"Program":[0,3]},
  specialized: [{term: 6, get: (value) => spec_Ident[value] || -1}],
  tokenPrec: 2802
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
                'for while loop if else match try catch return break continue impl': highlight.tags.controlKeyword,
                'in new': highlight.tags.operatorKeyword,
                'let fn type templ trigger': highlight.tags.definitionKeyword,
                'pub event': highlight.tags.modifier,
                'import from': highlight.tags.moduleKeyword,
                'ImportItem/Module/Ident': highlight.tags.namespace,
                // 'ImportItem/ImportItemCanditate/Ident': t.namespace,
                'ImportItem/ImportItemCanditate/TypeImport/Ident': highlight.tags.typeName,
                'ImportItem/ImportItemCanditate/TemplImport/Ident': highlight.tags.namespace,
                'ImportItem/ImportItemCanditate/TriggerImport/Ident': highlight.tags.local(highlight.tags.variableName),
                "Annotation": highlight.tags.separator,
                "TriggerConnective": highlight.tags.operatorKeyword,
                "ImplTemplateIdent/Ident": highlight.tags.namespace,
                "SingletonIdent/Ident": highlight.tags.typeName,
                'TypeDefinition/Ident': highlight.tags.namespace,
                'FunctionDefinition/Ident': highlight.tags.function(highlight.tags.variableName),
                'ObjectTypeFieldKey/Ident': highlight.tags.propertyName,
                'ObjectTypeFieldAnnotation ObjectTypeFieldAnnotation/Ident': highlight.tags.separator,
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
