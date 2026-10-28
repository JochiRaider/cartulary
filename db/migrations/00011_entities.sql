-- +goose Up
-- PostgreSQL 18.6 consolidated baseline: entities.
-- Administrator provisioning and final admission are separate lifecycle steps.

-- +goose StatementBegin
CREATE FUNCTION public.entities_source_codepoints_admitted_v1(candidate text, reject_format boolean)
 RETURNS boolean
 LANGUAGE plpgsql
 IMMUTABLE PARALLEL SAFE STRICT
 SET search_path = pg_catalog, public
AS $$
DECLARE
    codepoint integer;
    character_ordinal integer;
BEGIN
    FOR character_ordinal IN 1..char_length(candidate) LOOP
        codepoint := ascii(substring(candidate FROM character_ordinal FOR 1));
        IF codepoint BETWEEN 0 AND 31
            OR codepoint BETWEEN 127 AND 159
            OR (reject_format AND (
                codepoint = 173
                OR codepoint BETWEEN 1536 AND 1541
                OR codepoint = 1564
                OR codepoint = 1757
                OR codepoint = 1807
                OR codepoint BETWEEN 2192 AND 2193
                OR codepoint = 2274
                OR codepoint = 6158
                OR codepoint BETWEEN 8203 AND 8207
                OR codepoint BETWEEN 8234 AND 8238
                OR codepoint BETWEEN 8288 AND 8292
                OR codepoint BETWEEN 8294 AND 8303
                OR codepoint = 65279
                OR codepoint BETWEEN 65529 AND 65531
                OR codepoint IN (69821, 69837)
                OR codepoint BETWEEN 78896 AND 78911
                OR codepoint BETWEEN 113824 AND 113827
                OR codepoint BETWEEN 119155 AND 119162
                OR codepoint = 917505
                OR codepoint BETWEEN 917536 AND 917631
            )) THEN
            RETURN false;
        END IF;
    END LOOP;
    RETURN true;
END;
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.entities_source_codepoints_admitted_v1(candidate text, reject_format boolean) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION public.entities_source_codepoints_admitted_v1(candidate text, reject_format boolean) TO cartulary_runtime;

GRANT EXECUTE ON FUNCTION public.entities_source_codepoints_admitted_v1(candidate text, reject_format boolean) TO cartulary_recovery;

-- +goose StatementBegin
CREATE FUNCTION public.entities_trim_unicode_space_v1(candidate text)
 RETURNS text
 LANGUAGE sql
 IMMUTABLE PARALLEL SAFE STRICT
 SET search_path = pg_catalog, public
RETURN btrim(candidate, ((((((((((((((((((((((((chr(9) || chr(10)) || chr(11)) || chr(12)) || chr(13)) || chr(32)) || chr(133)) || chr(160)) || chr(5760)) || chr(8192)) || chr(8193)) || chr(8194)) || chr(8195)) || chr(8196)) || chr(8197)) || chr(8198)) || chr(8199)) || chr(8200)) || chr(8201)) || chr(8202)) || chr(8232)) || chr(8233)) || chr(8239)) || chr(8287)) || chr(12288)));
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.entities_trim_unicode_space_v1(candidate text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION public.entities_trim_unicode_space_v1(candidate text) TO cartulary_runtime;

GRANT EXECUTE ON FUNCTION public.entities_trim_unicode_space_v1(candidate text) TO cartulary_recovery;

-- +goose StatementBegin
CREATE FUNCTION public.entities_normalize_identifier_v1(identifier_type text, raw_value text)
 RETURNS text
 LANGUAGE plpgsql
 IMMUTABLE PARALLEL SAFE
 SET search_path = pg_catalog, public
AS $$
DECLARE
    normalized_value text;
BEGIN
    IF raw_value IS NULL OR identifier_type NOT IN (
        'aad_device_id', 'fqdn', 'hostname', 'aad_object_id',
        'sid', 'upn', 'email', 'sam_account_name'
    ) THEN
        RETURN NULL;
    END IF;
    normalized_value := normalize(
        public.entities_trim_unicode_space_v1(raw_value),
        NFC
    );
    IF normalized_value = ''
        OR NOT public.entities_source_codepoints_admitted_v1(
            normalized_value,
            true
        ) THEN
        RETURN NULL;
    END IF;
    IF identifier_type = 'sid' THEN
        RETURN translate(normalized_value, 'abcdefghijklmnopqrstuvwxyzµßàáâãäåæçèéêëìíîïðñòóôõöøùúûüýþÿāăąćĉċčďđēĕėęěĝğġģĥħĩīĭįıĳĵķĺļľŀłńņňŋōŏőœŕŗřśŝşšţťŧũūŭůűųŵŷźżžſƀƃƅƈƌƒƕƙƚƛƞơƣƥƨƭưƴƶƹƽƿǅǆǈǉǋǌǎǐǒǔǖǘǚǜǝǟǡǣǥǧǩǫǭǯǲǳǵǹǻǽǿȁȃȅȇȉȋȍȏȑȓȕȗșțȝȟȣȥȧȩȫȭȯȱȳȼȿɀɂɇɉɋɍɏɐɑɒɓɔɖɗəɛɜɠɡɣɤɥɦɨɩɪɫɬɯɱɲɵɽʀʂʃʇʈʉʊʋʌʒʝʞͅͱͳͷͻͼͽάέήίαβγδεζηθικλμνξοπρςστυφχψωϊϋόύώϐϑϕϖϗϙϛϝϟϡϣϥϧϩϫϭϯϰϱϲϳϵϸϻабвгдежзийклмнопрстуфхцчшщъыьэюяѐёђѓєѕіїјљњћќѝўџѡѣѥѧѩѫѭѯѱѳѵѷѹѻѽѿҁҋҍҏґғҕҗҙқҝҟҡңҥҧҩҫҭүұҳҵҷҹһҽҿӂӄӆӈӊӌӎӏӑӓӕӗәӛӝӟӡӣӥӧөӫӭӯӱӳӵӷӹӻӽӿԁԃԅԇԉԋԍԏԑԓԕԗԙԛԝԟԡԣԥԧԩԫԭԯաբգդեզէըթժիլխծկհձղճմյնշոչպջռսվտրցւփքօֆაბგდევზთიკლმნოპჟრსტუფქღყშჩცძწჭხჯჰჱჲჳჴჵჶჷჸჹჺჽჾჿᏸᏹᏺᏻᏼᏽᲀᲁᲂᲃᲄᲅᲆᲇᲈᲊᵹᵽᶎḁḃḅḇḉḋḍḏḑḓḕḗḙḛḝḟḡḣḥḧḩḫḭḯḱḳḵḷḹḻḽḿṁṃṅṇṉṋṍṏṑṓṕṗṙṛṝṟṡṣṥṧṩṫṭṯṱṳṵṷṹṻṽṿẁẃẅẇẉẋẍẏẑẓẕẛạảấầẩẫậắằẳẵặẹẻẽếềểễệỉịọỏốồổỗộớờởỡợụủứừửữựỳỵỷỹỻỽỿἀἁἂἃἄἅἆἇἐἑἒἓἔἕἠἡἢἣἤἥἦἧἰἱἲἳἴἵἶἷὀὁὂὃὄὅὑὓὕὗὠὡὢὣὤὥὦὧὰάὲέὴήὶίὸόὺύὼώᾀᾁᾂᾃᾄᾅᾆᾇᾐᾑᾒᾓᾔᾕᾖᾗᾠᾡᾢᾣᾤᾥᾦᾧᾰᾱᾳιῃῐῑῠῡῥῳⅎⅰⅱⅲⅳⅴⅵⅶⅷⅸⅹⅺⅻⅼⅽⅾⅿↄⓐⓑⓒⓓⓔⓕⓖⓗⓘⓙⓚⓛⓜⓝⓞⓟⓠⓡⓢⓣⓤⓥⓦⓧⓨⓩⰰⰱⰲⰳⰴⰵⰶⰷⰸⰹⰺⰻⰼⰽⰾⰿⱀⱁⱂⱃⱄⱅⱆⱇⱈⱉⱊⱋⱌⱍⱎⱏⱐⱑⱒⱓⱔⱕⱖⱗⱘⱙⱚⱛⱜⱝⱞⱟⱡⱥⱦⱨⱪⱬⱳⱶⲁⲃⲅⲇⲉⲋⲍⲏⲑⲓⲕⲗⲙⲛⲝⲟⲡⲣⲥⲧⲩⲫⲭⲯⲱⲳⲵⲷⲹⲻⲽⲿⳁⳃⳅⳇⳉⳋⳍⳏⳑⳓⳕⳗⳙⳛⳝⳟⳡⳣⳬⳮⳳⴀⴁⴂⴃⴄⴅⴆⴇⴈⴉⴊⴋⴌⴍⴎⴏⴐⴑⴒⴓⴔⴕⴖⴗⴘⴙⴚⴛⴜⴝⴞⴟⴠⴡⴢⴣⴤⴥⴧⴭꙁꙃꙅꙇꙉꙋꙍꙏꙑꙓꙕꙗꙙꙛꙝꙟꙡꙣꙥꙧꙩꙫꙭꚁꚃꚅꚇꚉꚋꚍꚏꚑꚓꚕꚗꚙꚛꜣꜥꜧꜩꜫꜭꜯꜳꜵꜷꜹꜻꜽꜿꝁꝃꝅꝇꝉꝋꝍꝏꝑꝓꝕꝗꝙꝛꝝꝟꝡꝣꝥꝧꝩꝫꝭꝯꝺꝼꝿꞁꞃꞅꞇꞌꞑꞓꞔꞗꞙꞛꞝꞟꞡꞣꞥꞧꞩꞵꞷꞹꞻꞽꞿꟁꟃꟈꟊꟍꟑꟗꟙꟛꟶꭓꭰꭱꭲꭳꭴꭵꭶꭷꭸꭹꭺꭻꭼꭽꭾꭿꮀꮁꮂꮃꮄꮅꮆꮇꮈꮉꮊꮋꮌꮍꮎꮏꮐꮑꮒꮓꮔꮕꮖꮗꮘꮙꮚꮛꮜꮝꮞꮟꮠꮡꮢꮣꮤꮥꮦꮧꮨꮩꮪꮫꮬꮭꮮꮯꮰꮱꮲꮳꮴꮵꮶꮷꮸꮹꮺꮻꮼꮽꮾꮿａｂｃｄｅｆｇｈｉｊｋｌｍｎｏｐｑｒｓｔｕｖｗｘｙｚ𐐨𐐩𐐪𐐫𐐬𐐭𐐮𐐯𐐰𐐱𐐲𐐳𐐴𐐵𐐶𐐷𐐸𐐹𐐺𐐻𐐼𐐽𐐾𐐿𐑀𐑁𐑂𐑃𐑄𐑅𐑆𐑇𐑈𐑉𐑊𐑋𐑌𐑍𐑎𐑏𐓘𐓙𐓚𐓛𐓜𐓝𐓞𐓟𐓠𐓡𐓢𐓣𐓤𐓥𐓦𐓧𐓨𐓩𐓪𐓫𐓬𐓭𐓮𐓯𐓰𐓱𐓲𐓳𐓴𐓵𐓶𐓷𐓸𐓹𐓺𐓻𐖗𐖘𐖙𐖚𐖛𐖜𐖝𐖞𐖟𐖠𐖡𐖣𐖤𐖥𐖦𐖧𐖨𐖩𐖪𐖫𐖬𐖭𐖮𐖯𐖰𐖱𐖳𐖴𐖵𐖶𐖷𐖸𐖹𐖻𐖼𐳀𐳁𐳂𐳃𐳄𐳅𐳆𐳇𐳈𐳉𐳊𐳋𐳌𐳍𐳎𐳏𐳐𐳑𐳒𐳓𐳔𐳕𐳖𐳗𐳘𐳙𐳚𐳛𐳜𐳝𐳞𐳟𐳠𐳡𐳢𐳣𐳤𐳥𐳦𐳧𐳨𐳩𐳪𐳫𐳬𐳭𐳮𐳯𐳰𐳱𐳲𐵰𐵱𐵲𐵳𐵴𐵵𐵶𐵷𐵸𐵹𐵺𐵻𐵼𐵽𐵾𐵿𐶀𐶁𐶂𐶃𐶄𐶅𑣀𑣁𑣂𑣃𑣄𑣅𑣆𑣇𑣈𑣉𑣊𑣋𑣌𑣍𑣎𑣏𑣐𑣑𑣒𑣓𑣔𑣕𑣖𑣗𑣘𑣙𑣚𑣛𑣜𑣝𑣞𑣟𖹠𖹡𖹢𖹣𖹤𖹥𖹦𖹧𖹨𖹩𖹪𖹫𖹬𖹭𖹮𖹯𖹰𖹱𖹲𖹳𖹴𖹵𖹶𖹷𖹸𖹹𖹺𖹻𖹼𖹽𖹾𖹿𞤢𞤣𞤤𞤥𞤦𞤧𞤨𞤩𞤪𞤫𞤬𞤭𞤮𞤯𞤰𞤱𞤲𞤳𞤴𞤵𞤶𞤷𞤸𞤹𞤺𞤻𞤼𞤽𞤾𞤿𞥀𞥁𞥂𞥃', 'ABCDEFGHIJKLMNOPQRSTUVWXYZΜẞÀÁÂÃÄÅÆÇÈÉÊËÌÍÎÏÐÑÒÓÔÕÖØÙÚÛÜÝÞŸĀĂĄĆĈĊČĎĐĒĔĖĘĚĜĞĠĢĤĦĨĪĬĮIĲĴĶĹĻĽĿŁŃŅŇŊŌŎŐŒŔŖŘŚŜŞŠŢŤŦŨŪŬŮŰŲŴŶŹŻŽSɃƂƄƇƋƑǶƘȽꟜȠƠƢƤƧƬƯƳƵƸƼǷǄǄǇǇǊǊǍǏǑǓǕǗǙǛƎǞǠǢǤǦǨǪǬǮǱǱǴǸǺǼǾȀȂȄȆȈȊȌȎȐȒȔȖȘȚȜȞȢȤȦȨȪȬȮȰȲȻⱾⱿɁɆɈɊɌɎⱯⱭⱰƁƆƉƊƏƐꞫƓꞬƔꟋꞍꞪƗƖꞮⱢꞭƜⱮƝƟⱤƦꟅƩꞱƮɄƱƲɅƷꞲꞰΙͰͲͶϽϾϿΆΈΉΊΑΒΓΔΕΖΗΘΙΚΛΜΝΞΟΠΡΣΣΤΥΦΧΨΩΪΫΌΎΏΒΘΦΠϏϘϚϜϞϠϢϤϦϨϪϬϮΚΡϹͿΕϷϺАБВГДЕЖЗИЙКЛМНОПРСТУФХЦЧШЩЪЫЬЭЮЯЀЁЂЃЄЅІЇЈЉЊЋЌЍЎЏѠѢѤѦѨѪѬѮѰѲѴѶѸѺѼѾҀҊҌҎҐҒҔҖҘҚҜҞҠҢҤҦҨҪҬҮҰҲҴҶҸҺҼҾӁӃӅӇӉӋӍӀӐӒӔӖӘӚӜӞӠӢӤӦӨӪӬӮӰӲӴӶӸӺӼӾԀԂԄԆԈԊԌԎԐԒԔԖԘԚԜԞԠԢԤԦԨԪԬԮԱԲԳԴԵԶԷԸԹԺԻԼԽԾԿՀՁՂՃՄՅՆՇՈՉՊՋՌՍՎՏՐՑՒՓՔՕՖᲐᲑᲒᲓᲔᲕᲖᲗᲘᲙᲚᲛᲜᲝᲞᲟᲠᲡᲢᲣᲤᲥᲦᲧᲨᲩᲪᲫᲬᲭᲮᲯᲰᲱᲲᲳᲴᲵᲶᲷᲸᲹᲺᲽᲾᲿᏰᏱᏲᏳᏴᏵВДОСТТЪѢꙊᲉꝽⱣꟆḀḂḄḆḈḊḌḎḐḒḔḖḘḚḜḞḠḢḤḦḨḪḬḮḰḲḴḶḸḺḼḾṀṂṄṆṈṊṌṎṐṒṔṖṘṚṜṞṠṢṤṦṨṪṬṮṰṲṴṶṸṺṼṾẀẂẄẆẈẊẌẎẐẒẔṠẠẢẤẦẨẪẬẮẰẲẴẶẸẺẼẾỀỂỄỆỈỊỌỎỐỒỔỖỘỚỜỞỠỢỤỦỨỪỬỮỰỲỴỶỸỺỼỾἈἉἊἋἌἍἎἏἘἙἚἛἜἝἨἩἪἫἬἭἮἯἸἹἺἻἼἽἾἿὈὉὊὋὌὍὙὛὝὟὨὩὪὫὬὭὮὯᾺΆῈΈῊΉῚΊῸΌῪΎῺΏᾈᾉᾊᾋᾌᾍᾎᾏᾘᾙᾚᾛᾜᾝᾞᾟᾨᾩᾪᾫᾬᾭᾮᾯᾸᾹᾼΙῌῘῙῨῩῬῼℲⅠⅡⅢⅣⅤⅥⅦⅧⅨⅩⅪⅫⅬⅭⅮⅯↃⒶⒷⒸⒹⒺⒻⒼⒽⒾⒿⓀⓁⓂⓃⓄⓅⓆⓇⓈⓉⓊⓋⓌⓍⓎⓏⰀⰁⰂⰃⰄⰅⰆⰇⰈⰉⰊⰋⰌⰍⰎⰏⰐⰑⰒⰓⰔⰕⰖⰗⰘⰙⰚⰛⰜⰝⰞⰟⰠⰡⰢⰣⰤⰥⰦⰧⰨⰩⰪⰫⰬⰭⰮⰯⱠȺȾⱧⱩⱫⱲⱵⲀⲂⲄⲆⲈⲊⲌⲎⲐⲒⲔⲖⲘⲚⲜⲞⲠⲢⲤⲦⲨⲪⲬⲮⲰⲲⲴⲶⲸⲺⲼⲾⳀⳂⳄⳆⳈⳊⳌⳎⳐⳒⳔⳖⳘⳚⳜⳞⳠⳢⳫⳭⳲႠႡႢႣႤႥႦႧႨႩႪႫႬႭႮႯႰႱႲႳႴႵႶႷႸႹႺႻႼႽႾႿჀჁჂჃჄჅჇჍꙀꙂꙄꙆꙈꙊꙌꙎꙐꙒꙔꙖꙘꙚꙜꙞꙠꙢꙤꙦꙨꙪꙬꚀꚂꚄꚆꚈꚊꚌꚎꚐꚒꚔꚖꚘꚚꜢꜤꜦꜨꜪꜬꜮꜲꜴꜶꜸꜺꜼꜾꝀꝂꝄꝆꝈꝊꝌꝎꝐꝒꝔꝖꝘꝚꝜꝞꝠꝢꝤꝦꝨꝪꝬꝮꝹꝻꝾꞀꞂꞄꞆꞋꞐꞒꟄꞖꞘꞚꞜꞞꞠꞢꞤꞦꞨꞴꞶꞸꞺꞼꞾꟀꟂꟇꟉꟌꟐꟖꟘꟚꟵꞳᎠᎡᎢᎣᎤᎥᎦᎧᎨᎩᎪᎫᎬᎭᎮᎯᎰᎱᎲᎳᎴᎵᎶᎷᎸᎹᎺᎻᎼᎽᎾᎿᏀᏁᏂᏃᏄᏅᏆᏇᏈᏉᏊᏋᏌᏍᏎᏏᏐᏑᏒᏓᏔᏕᏖᏗᏘᏙᏚᏛᏜᏝᏞᏟᏠᏡᏢᏣᏤᏥᏦᏧᏨᏩᏪᏫᏬᏭᏮᏯＡＢＣＤＥＦＧＨＩＪＫＬＭＮＯＰＱＲＳＴＵＶＷＸＹＺ𐐀𐐁𐐂𐐃𐐄𐐅𐐆𐐇𐐈𐐉𐐊𐐋𐐌𐐍𐐎𐐏𐐐𐐑𐐒𐐓𐐔𐐕𐐖𐐗𐐘𐐙𐐚𐐛𐐜𐐝𐐞𐐟𐐠𐐡𐐢𐐣𐐤𐐥𐐦𐐧𐒰𐒱𐒲𐒳𐒴𐒵𐒶𐒷𐒸𐒹𐒺𐒻𐒼𐒽𐒾𐒿𐓀𐓁𐓂𐓃𐓄𐓅𐓆𐓇𐓈𐓉𐓊𐓋𐓌𐓍𐓎𐓏𐓐𐓑𐓒𐓓𐕰𐕱𐕲𐕳𐕴𐕵𐕶𐕷𐕸𐕹𐕺𐕼𐕽𐕾𐕿𐖀𐖁𐖂𐖃𐖄𐖅𐖆𐖇𐖈𐖉𐖊𐖌𐖍𐖎𐖏𐖐𐖑𐖒𐖔𐖕𐲀𐲁𐲂𐲃𐲄𐲅𐲆𐲇𐲈𐲉𐲊𐲋𐲌𐲍𐲎𐲏𐲐𐲑𐲒𐲓𐲔𐲕𐲖𐲗𐲘𐲙𐲚𐲛𐲜𐲝𐲞𐲟𐲠𐲡𐲢𐲣𐲤𐲥𐲦𐲧𐲨𐲩𐲪𐲫𐲬𐲭𐲮𐲯𐲰𐲱𐲲𐵐𐵑𐵒𐵓𐵔𐵕𐵖𐵗𐵘𐵙𐵚𐵛𐵜𐵝𐵞𐵟𐵠𐵡𐵢𐵣𐵤𐵥𑢠𑢡𑢢𑢣𑢤𑢥𑢦𑢧𑢨𑢩𑢪𑢫𑢬𑢭𑢮𑢯𑢰𑢱𑢲𑢳𑢴𑢵𑢶𑢷𑢸𑢹𑢺𑢻𑢼𑢽𑢾𑢿𖹀𖹁𖹂𖹃𖹄𖹅𖹆𖹇𖹈𖹉𖹊𖹋𖹌𖹍𖹎𖹏𖹐𖹑𖹒𖹓𖹔𖹕𖹖𖹗𖹘𖹙𖹚𖹛𖹜𖹝𖹞𖹟𞤀𞤁𞤂𞤃𞤄𞤅𞤆𞤇𞤈𞤉𞤊𞤋𞤌𞤍𞤎𞤏𞤐𞤑𞤒𞤓𞤔𞤕𞤖𞤗𞤘𞤙𞤚𞤛𞤜𞤝𞤞𞤟𞤠𞤡');
    END IF;
    RETURN translate(normalized_value, 'ABCDEFGHIJKLMNOPQRSTUVWXYZÀÁÂÃÄÅÆÇÈÉÊËÌÍÎÏÐÑÒÓÔÕÖØÙÚÛÜÝÞĀĂĄĆĈĊČĎĐĒĔĖĘĚĜĞĠĢĤĦĨĪĬĮİĲĴĶĹĻĽĿŁŃŅŇŊŌŎŐŒŔŖŘŚŜŞŠŢŤŦŨŪŬŮŰŲŴŶŸŹŻŽƁƂƄƆƇƉƊƋƎƏƐƑƓƔƖƗƘƜƝƟƠƢƤƦƧƩƬƮƯƱƲƳƵƷƸƼǄǅǇǈǊǋǍǏǑǓǕǗǙǛǞǠǢǤǦǨǪǬǮǱǲǴǶǷǸǺǼǾȀȂȄȆȈȊȌȎȐȒȔȖȘȚȜȞȠȢȤȦȨȪȬȮȰȲȺȻȽȾɁɃɄɅɆɈɊɌɎͰͲͶͿΆΈΉΊΌΎΏΑΒΓΔΕΖΗΘΙΚΛΜΝΞΟΠΡΣΤΥΦΧΨΩΪΫϏϘϚϜϞϠϢϤϦϨϪϬϮϴϷϹϺϽϾϿЀЁЂЃЄЅІЇЈЉЊЋЌЍЎЏАБВГДЕЖЗИЙКЛМНОПРСТУФХЦЧШЩЪЫЬЭЮЯѠѢѤѦѨѪѬѮѰѲѴѶѸѺѼѾҀҊҌҎҐҒҔҖҘҚҜҞҠҢҤҦҨҪҬҮҰҲҴҶҸҺҼҾӀӁӃӅӇӉӋӍӐӒӔӖӘӚӜӞӠӢӤӦӨӪӬӮӰӲӴӶӸӺӼӾԀԂԄԆԈԊԌԎԐԒԔԖԘԚԜԞԠԢԤԦԨԪԬԮԱԲԳԴԵԶԷԸԹԺԻԼԽԾԿՀՁՂՃՄՅՆՇՈՉՊՋՌՍՎՏՐՑՒՓՔՕՖႠႡႢႣႤႥႦႧႨႩႪႫႬႭႮႯႰႱႲႳႴႵႶႷႸႹႺႻႼႽႾႿჀჁჂჃჄჅჇჍᎠᎡᎢᎣᎤᎥᎦᎧᎨᎩᎪᎫᎬᎭᎮᎯᎰᎱᎲᎳᎴᎵᎶᎷᎸᎹᎺᎻᎼᎽᎾᎿᏀᏁᏂᏃᏄᏅᏆᏇᏈᏉᏊᏋᏌᏍᏎᏏᏐᏑᏒᏓᏔᏕᏖᏗᏘᏙᏚᏛᏜᏝᏞᏟᏠᏡᏢᏣᏤᏥᏦᏧᏨᏩᏪᏫᏬᏭᏮᏯᏰᏱᏲᏳᏴᏵᲉᲐᲑᲒᲓᲔᲕᲖᲗᲘᲙᲚᲛᲜᲝᲞᲟᲠᲡᲢᲣᲤᲥᲦᲧᲨᲩᲪᲫᲬᲭᲮᲯᲰᲱᲲᲳᲴᲵᲶᲷᲸᲹᲺᲽᲾᲿḀḂḄḆḈḊḌḎḐḒḔḖḘḚḜḞḠḢḤḦḨḪḬḮḰḲḴḶḸḺḼḾṀṂṄṆṈṊṌṎṐṒṔṖṘṚṜṞṠṢṤṦṨṪṬṮṰṲṴṶṸṺṼṾẀẂẄẆẈẊẌẎẐẒẔẞẠẢẤẦẨẪẬẮẰẲẴẶẸẺẼẾỀỂỄỆỈỊỌỎỐỒỔỖỘỚỜỞỠỢỤỦỨỪỬỮỰỲỴỶỸỺỼỾἈἉἊἋἌἍἎἏἘἙἚἛἜἝἨἩἪἫἬἭἮἯἸἹἺἻἼἽἾἿὈὉὊὋὌὍὙὛὝὟὨὩὪὫὬὭὮὯᾈᾉᾊᾋᾌᾍᾎᾏᾘᾙᾚᾛᾜᾝᾞᾟᾨᾩᾪᾫᾬᾭᾮᾯᾸᾹᾺΆᾼῈΈῊΉῌῘῙῚΊῨῩῪΎῬῸΌῺΏῼΩKÅℲⅠⅡⅢⅣⅤⅥⅦⅧⅨⅩⅪⅫⅬⅭⅮⅯↃⒶⒷⒸⒹⒺⒻⒼⒽⒾⒿⓀⓁⓂⓃⓄⓅⓆⓇⓈⓉⓊⓋⓌⓍⓎⓏⰀⰁⰂⰃⰄⰅⰆⰇⰈⰉⰊⰋⰌⰍⰎⰏⰐⰑⰒⰓⰔⰕⰖⰗⰘⰙⰚⰛⰜⰝⰞⰟⰠⰡⰢⰣⰤⰥⰦⰧⰨⰩⰪⰫⰬⰭⰮⰯⱠⱢⱣⱤⱧⱩⱫⱭⱮⱯⱰⱲⱵⱾⱿⲀⲂⲄⲆⲈⲊⲌⲎⲐⲒⲔⲖⲘⲚⲜⲞⲠⲢⲤⲦⲨⲪⲬⲮⲰⲲⲴⲶⲸⲺⲼⲾⳀⳂⳄⳆⳈⳊⳌⳎⳐⳒⳔⳖⳘⳚⳜⳞⳠⳢⳫⳭⳲꙀꙂꙄꙆꙈꙊꙌꙎꙐꙒꙔꙖꙘꙚꙜꙞꙠꙢꙤꙦꙨꙪꙬꚀꚂꚄꚆꚈꚊꚌꚎꚐꚒꚔꚖꚘꚚꜢꜤꜦꜨꜪꜬꜮꜲꜴꜶꜸꜺꜼꜾꝀꝂꝄꝆꝈꝊꝌꝎꝐꝒꝔꝖꝘꝚꝜꝞꝠꝢꝤꝦꝨꝪꝬꝮꝹꝻꝽꝾꞀꞂꞄꞆꞋꞍꞐꞒꞖꞘꞚꞜꞞꞠꞢꞤꞦꞨꞪꞫꞬꞭꞮꞰꞱꞲꞳꞴꞶꞸꞺꞼꞾꟀꟂꟄꟅꟆꟇꟉꟋꟌꟐꟖꟘꟚꟜꟵＡＢＣＤＥＦＧＨＩＪＫＬＭＮＯＰＱＲＳＴＵＶＷＸＹＺ𐐀𐐁𐐂𐐃𐐄𐐅𐐆𐐇𐐈𐐉𐐊𐐋𐐌𐐍𐐎𐐏𐐐𐐑𐐒𐐓𐐔𐐕𐐖𐐗𐐘𐐙𐐚𐐛𐐜𐐝𐐞𐐟𐐠𐐡𐐢𐐣𐐤𐐥𐐦𐐧𐒰𐒱𐒲𐒳𐒴𐒵𐒶𐒷𐒸𐒹𐒺𐒻𐒼𐒽𐒾𐒿𐓀𐓁𐓂𐓃𐓄𐓅𐓆𐓇𐓈𐓉𐓊𐓋𐓌𐓍𐓎𐓏𐓐𐓑𐓒𐓓𐕰𐕱𐕲𐕳𐕴𐕵𐕶𐕷𐕸𐕹𐕺𐕼𐕽𐕾𐕿𐖀𐖁𐖂𐖃𐖄𐖅𐖆𐖇𐖈𐖉𐖊𐖌𐖍𐖎𐖏𐖐𐖑𐖒𐖔𐖕𐲀𐲁𐲂𐲃𐲄𐲅𐲆𐲇𐲈𐲉𐲊𐲋𐲌𐲍𐲎𐲏𐲐𐲑𐲒𐲓𐲔𐲕𐲖𐲗𐲘𐲙𐲚𐲛𐲜𐲝𐲞𐲟𐲠𐲡𐲢𐲣𐲤𐲥𐲦𐲧𐲨𐲩𐲪𐲫𐲬𐲭𐲮𐲯𐲰𐲱𐲲𐵐𐵑𐵒𐵓𐵔𐵕𐵖𐵗𐵘𐵙𐵚𐵛𐵜𐵝𐵞𐵟𐵠𐵡𐵢𐵣𐵤𐵥𑢠𑢡𑢢𑢣𑢤𑢥𑢦𑢧𑢨𑢩𑢪𑢫𑢬𑢭𑢮𑢯𑢰𑢱𑢲𑢳𑢴𑢵𑢶𑢷𑢸𑢹𑢺𑢻𑢼𑢽𑢾𑢿𖹀𖹁𖹂𖹃𖹄𖹅𖹆𖹇𖹈𖹉𖹊𖹋𖹌𖹍𖹎𖹏𖹐𖹑𖹒𖹓𖹔𖹕𖹖𖹗𖹘𖹙𖹚𖹛𖹜𖹝𖹞𖹟𞤀𞤁𞤂𞤃𞤄𞤅𞤆𞤇𞤈𞤉𞤊𞤋𞤌𞤍𞤎𞤏𞤐𞤑𞤒𞤓𞤔𞤕𞤖𞤗𞤘𞤙𞤚𞤛𞤜𞤝𞤞𞤟𞤠𞤡', 'abcdefghijklmnopqrstuvwxyzàáâãäåæçèéêëìíîïðñòóôõöøùúûüýþāăąćĉċčďđēĕėęěĝğġģĥħĩīĭįiĳĵķĺļľŀłńņňŋōŏőœŕŗřśŝşšţťŧũūŭůűųŵŷÿźżžɓƃƅɔƈɖɗƌǝəɛƒɠɣɩɨƙɯɲɵơƣƥʀƨʃƭʈưʊʋƴƶʒƹƽǆǆǉǉǌǌǎǐǒǔǖǘǚǜǟǡǣǥǧǩǫǭǯǳǳǵƕƿǹǻǽǿȁȃȅȇȉȋȍȏȑȓȕȗșțȝȟƞȣȥȧȩȫȭȯȱȳⱥȼƚⱦɂƀʉʌɇɉɋɍɏͱͳͷϳάέήίόύώαβγδεζηθικλμνξοπρστυφχψωϊϋϗϙϛϝϟϡϣϥϧϩϫϭϯθϸϲϻͻͼͽѐёђѓєѕіїјљњћќѝўџабвгдежзийклмнопрстуфхцчшщъыьэюяѡѣѥѧѩѫѭѯѱѳѵѷѹѻѽѿҁҋҍҏґғҕҗҙқҝҟҡңҥҧҩҫҭүұҳҵҷҹһҽҿӏӂӄӆӈӊӌӎӑӓӕӗәӛӝӟӡӣӥӧөӫӭӯӱӳӵӷӹӻӽӿԁԃԅԇԉԋԍԏԑԓԕԗԙԛԝԟԡԣԥԧԩԫԭԯաբգդեզէըթժիլխծկհձղճմյնշոչպջռսվտրցւփքօֆⴀⴁⴂⴃⴄⴅⴆⴇⴈⴉⴊⴋⴌⴍⴎⴏⴐⴑⴒⴓⴔⴕⴖⴗⴘⴙⴚⴛⴜⴝⴞⴟⴠⴡⴢⴣⴤⴥⴧⴭꭰꭱꭲꭳꭴꭵꭶꭷꭸꭹꭺꭻꭼꭽꭾꭿꮀꮁꮂꮃꮄꮅꮆꮇꮈꮉꮊꮋꮌꮍꮎꮏꮐꮑꮒꮓꮔꮕꮖꮗꮘꮙꮚꮛꮜꮝꮞꮟꮠꮡꮢꮣꮤꮥꮦꮧꮨꮩꮪꮫꮬꮭꮮꮯꮰꮱꮲꮳꮴꮵꮶꮷꮸꮹꮺꮻꮼꮽꮾꮿᏸᏹᏺᏻᏼᏽᲊაბგდევზთიკლმნოპჟრსტუფქღყშჩცძწჭხჯჰჱჲჳჴჵჶჷჸჹჺჽჾჿḁḃḅḇḉḋḍḏḑḓḕḗḙḛḝḟḡḣḥḧḩḫḭḯḱḳḵḷḹḻḽḿṁṃṅṇṉṋṍṏṑṓṕṗṙṛṝṟṡṣṥṧṩṫṭṯṱṳṵṷṹṻṽṿẁẃẅẇẉẋẍẏẑẓẕßạảấầẩẫậắằẳẵặẹẻẽếềểễệỉịọỏốồổỗộớờởỡợụủứừửữựỳỵỷỹỻỽỿἀἁἂἃἄἅἆἇἐἑἒἓἔἕἠἡἢἣἤἥἦἧἰἱἲἳἴἵἶἷὀὁὂὃὄὅὑὓὕὗὠὡὢὣὤὥὦὧᾀᾁᾂᾃᾄᾅᾆᾇᾐᾑᾒᾓᾔᾕᾖᾗᾠᾡᾢᾣᾤᾥᾦᾧᾰᾱὰάᾳὲέὴήῃῐῑὶίῠῡὺύῥὸόὼώῳωkåⅎⅰⅱⅲⅳⅴⅵⅶⅷⅸⅹⅺⅻⅼⅽⅾⅿↄⓐⓑⓒⓓⓔⓕⓖⓗⓘⓙⓚⓛⓜⓝⓞⓟⓠⓡⓢⓣⓤⓥⓦⓧⓨⓩⰰⰱⰲⰳⰴⰵⰶⰷⰸⰹⰺⰻⰼⰽⰾⰿⱀⱁⱂⱃⱄⱅⱆⱇⱈⱉⱊⱋⱌⱍⱎⱏⱐⱑⱒⱓⱔⱕⱖⱗⱘⱙⱚⱛⱜⱝⱞⱟⱡɫᵽɽⱨⱪⱬɑɱɐɒⱳⱶȿɀⲁⲃⲅⲇⲉⲋⲍⲏⲑⲓⲕⲗⲙⲛⲝⲟⲡⲣⲥⲧⲩⲫⲭⲯⲱⲳⲵⲷⲹⲻⲽⲿⳁⳃⳅⳇⳉⳋⳍⳏⳑⳓⳕⳗⳙⳛⳝⳟⳡⳣⳬⳮⳳꙁꙃꙅꙇꙉꙋꙍꙏꙑꙓꙕꙗꙙꙛꙝꙟꙡꙣꙥꙧꙩꙫꙭꚁꚃꚅꚇꚉꚋꚍꚏꚑꚓꚕꚗꚙꚛꜣꜥꜧꜩꜫꜭꜯꜳꜵꜷꜹꜻꜽꜿꝁꝃꝅꝇꝉꝋꝍꝏꝑꝓꝕꝗꝙꝛꝝꝟꝡꝣꝥꝧꝩꝫꝭꝯꝺꝼᵹꝿꞁꞃꞅꞇꞌɥꞑꞓꞗꞙꞛꞝꞟꞡꞣꞥꞧꞩɦɜɡɬɪʞʇʝꭓꞵꞷꞹꞻꞽꞿꟁꟃꞔʂᶎꟈꟊɤꟍꟑꟗꟙꟛƛꟶａｂｃｄｅｆｇｈｉｊｋｌｍｎｏｐｑｒｓｔｕｖｗｘｙｚ𐐨𐐩𐐪𐐫𐐬𐐭𐐮𐐯𐐰𐐱𐐲𐐳𐐴𐐵𐐶𐐷𐐸𐐹𐐺𐐻𐐼𐐽𐐾𐐿𐑀𐑁𐑂𐑃𐑄𐑅𐑆𐑇𐑈𐑉𐑊𐑋𐑌𐑍𐑎𐑏𐓘𐓙𐓚𐓛𐓜𐓝𐓞𐓟𐓠𐓡𐓢𐓣𐓤𐓥𐓦𐓧𐓨𐓩𐓪𐓫𐓬𐓭𐓮𐓯𐓰𐓱𐓲𐓳𐓴𐓵𐓶𐓷𐓸𐓹𐓺𐓻𐖗𐖘𐖙𐖚𐖛𐖜𐖝𐖞𐖟𐖠𐖡𐖣𐖤𐖥𐖦𐖧𐖨𐖩𐖪𐖫𐖬𐖭𐖮𐖯𐖰𐖱𐖳𐖴𐖵𐖶𐖷𐖸𐖹𐖻𐖼𐳀𐳁𐳂𐳃𐳄𐳅𐳆𐳇𐳈𐳉𐳊𐳋𐳌𐳍𐳎𐳏𐳐𐳑𐳒𐳓𐳔𐳕𐳖𐳗𐳘𐳙𐳚𐳛𐳜𐳝𐳞𐳟𐳠𐳡𐳢𐳣𐳤𐳥𐳦𐳧𐳨𐳩𐳪𐳫𐳬𐳭𐳮𐳯𐳰𐳱𐳲𐵰𐵱𐵲𐵳𐵴𐵵𐵶𐵷𐵸𐵹𐵺𐵻𐵼𐵽𐵾𐵿𐶀𐶁𐶂𐶃𐶄𐶅𑣀𑣁𑣂𑣃𑣄𑣅𑣆𑣇𑣈𑣉𑣊𑣋𑣌𑣍𑣎𑣏𑣐𑣑𑣒𑣓𑣔𑣕𑣖𑣗𑣘𑣙𑣚𑣛𑣜𑣝𑣞𑣟𖹠𖹡𖹢𖹣𖹤𖹥𖹦𖹧𖹨𖹩𖹪𖹫𖹬𖹭𖹮𖹯𖹰𖹱𖹲𖹳𖹴𖹵𖹶𖹷𖹸𖹹𖹺𖹻𖹼𖹽𖹾𖹿𞤢𞤣𞤤𞤥𞤦𞤧𞤨𞤩𞤪𞤫𞤬𞤭𞤮𞤯𞤰𞤱𞤲𞤳𞤴𞤵𞤶𞤷𞤸𞤹𞤺𞤻𞤼𞤽𞤾𞤿𞥀𞥁𞥂𞥃');
END;
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.entities_normalize_identifier_v1(identifier_type text, raw_value text) FROM PUBLIC;

CREATE TABLE public.entity_active_identifier_claims (
    incident_id uuid NOT NULL,
    entity_type text NOT NULL,
    identifier_type text NOT NULL,
    normalized_value text NOT NULL,
    record_id uuid NOT NULL,
    CONSTRAINT entity_active_identifier_claims_entity_type_ck CHECK (entity_type = ANY (ARRAY['host'::text, 'identity'::text])),
    CONSTRAINT entity_active_identifier_claims_identifier_class_ck CHECK (entity_type = 'host'::text AND (identifier_type = ANY (ARRAY['aad_device_id'::text, 'fqdn'::text, 'hostname'::text])) OR entity_type = 'identity'::text AND (identifier_type = ANY (ARRAY['aad_object_id'::text, 'sid'::text, 'upn'::text, 'email'::text, 'sam_account_name'::text]))),
    CONSTRAINT entity_active_identifier_claims_pkey PRIMARY KEY (incident_id, entity_type, identifier_type, normalized_value),
    CONSTRAINT entity_active_identifier_claims_value_ck CHECK (normalized_value <> ''::text AND public.entities_normalize_identifier_v1(identifier_type, normalized_value) = normalized_value)
);

REVOKE ALL ON TABLE public.entity_active_identifier_claims FROM PUBLIC;

REVOKE ALL ON TYPE public.entity_active_identifier_claims FROM PUBLIC;

GRANT SELECT ON TABLE public.entity_active_identifier_claims TO cartulary_runtime;

GRANT SELECT, TRUNCATE ON TABLE public.entity_active_identifier_claims TO cartulary_recovery;

CREATE TABLE public.entity_aliases (
    entity_alias_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    record_id uuid NOT NULL,
    entity_type text NOT NULL,
    raw_text text NOT NULL,
    normalized_text public.citext NOT NULL,
    classification text NOT NULL,
    created_by_user_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    CONSTRAINT entity_aliases_alias_text_controls_ck CHECK (normalized_text::text !~ '[[:cntrl:]]'::text),
    CONSTRAINT entity_aliases_alias_text_length_ck CHECK (char_length(normalized_text::text) <= 256),
    CONSTRAINT entity_aliases_alias_text_nonempty_ck CHECK (normalized_text::text <> ''::text),
    CONSTRAINT entity_aliases_alias_text_storage_ck CHECK (raw_text = normalized_text::text),
    CONSTRAINT entity_aliases_classification_check CHECK (classification = 'suggestion_only'::text),
    CONSTRAINT entity_aliases_entity_type_check CHECK (entity_type = ANY (ARRAY['host'::text, 'identity'::text])),
    CONSTRAINT entity_aliases_pkey PRIMARY KEY (entity_alias_id)
);

REVOKE ALL ON TABLE public.entity_aliases FROM PUBLIC;

REVOKE ALL ON TYPE public.entity_aliases FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.entity_aliases TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.entity_aliases TO cartulary_recovery;

CREATE TABLE public.entity_mentions (
    entity_mention_id uuid DEFAULT gen_random_uuid() NOT NULL,
    source_record_id uuid NOT NULL,
    entity_type text NOT NULL,
    source_field_key text NOT NULL,
    origin_kind text NOT NULL,
    origin_locator text NOT NULL,
    raw_text text NOT NULL,
    normalized_text text NOT NULL,
    resolution_status text NOT NULL,
    row_version bigint DEFAULT 1 NOT NULL,
    ordinal integer NOT NULL,
    created_by_user_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    resolved_record_id uuid,
    resolved_by_user_id uuid,
    resolved_at timestamp with time zone,
    resolution_method text,
    CONSTRAINT entity_mentions_entity_type_check CHECK (entity_type = ANY (ARRAY['host'::text, 'identity'::text])),
    CONSTRAINT entity_mentions_ordinal_check CHECK (ordinal > 0),
    CONSTRAINT entity_mentions_origin_kind_ck CHECK (origin_kind = ANY (ARRAY['manual_entry'::text, 'clipboard_paste'::text, 'csv_import'::text, 'xlsx_import'::text, 'api_import'::text, 'extraction'::text, 'system'::text])),
    CONSTRAINT entity_mentions_pkey PRIMARY KEY (entity_mention_id),
    CONSTRAINT entity_mentions_required_text_ck CHECK (btrim(source_field_key) <> ''::text AND btrim(origin_locator) <> ''::text AND btrim(raw_text) <> ''::text AND btrim(normalized_text) <> ''::text),
    CONSTRAINT entity_mentions_resolution_status_check CHECK (resolution_status = ANY (ARRAY['unresolved'::text, 'resolved'::text, 'dismissed'::text])),
    CONSTRAINT entity_mentions_resolution_tuple_ck CHECK (resolution_status = 'resolved'::text AND resolved_record_id IS NOT NULL AND resolved_by_user_id IS NOT NULL AND resolved_at IS NOT NULL AND btrim(COALESCE(resolution_method, ''::text)) <> ''::text OR (resolution_status = ANY (ARRAY['unresolved'::text, 'dismissed'::text])) AND resolved_record_id IS NULL AND resolved_by_user_id IS NULL AND resolved_at IS NULL AND resolution_method IS NULL),
    CONSTRAINT entity_mentions_row_version_ck CHECK (row_version >= 1)
);

REVOKE ALL ON TABLE public.entity_mentions FROM PUBLIC;

REVOKE ALL ON TYPE public.entity_mentions FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.entity_mentions TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.entity_mentions TO cartulary_recovery;

CREATE TABLE public.entity_preserved_identifiers (
    entity_preserved_identifier_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    record_id uuid NOT NULL,
    entity_type text NOT NULL,
    identifier_type text NOT NULL,
    raw_value text NOT NULL,
    normalized_value text NOT NULL,
    classification text NOT NULL,
    created_by_user_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    CONSTRAINT entity_preserved_identifiers_classification_check CHECK (classification = ANY (ARRAY['exact_match_reuse'::text, 'suggestion_only'::text, 'provenance_only'::text])),
    CONSTRAINT entity_preserved_identifiers_entity_type_check CHECK (entity_type = ANY (ARRAY['host'::text, 'identity'::text])),
    CONSTRAINT entity_preserved_identifiers_identifier_class_ck CHECK (entity_type = 'host'::text AND (identifier_type = ANY (ARRAY['aad_device_id'::text, 'fqdn'::text, 'hostname'::text])) OR entity_type = 'identity'::text AND (identifier_type = ANY (ARRAY['aad_object_id'::text, 'sid'::text, 'upn'::text, 'email'::text, 'sam_account_name'::text]))),
    CONSTRAINT entity_preserved_identifiers_pkey PRIMARY KEY (entity_preserved_identifier_id),
    CONSTRAINT entity_preserved_identifiers_value_nonempty_ck CHECK (btrim(raw_value) <> ''::text AND btrim(normalized_value) <> ''::text)
);

REVOKE ALL ON TABLE public.entity_preserved_identifiers FROM PUBLIC;

REVOKE ALL ON TYPE public.entity_preserved_identifiers FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.entity_preserved_identifiers TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.entity_preserved_identifiers TO cartulary_recovery;

CREATE TABLE public.hosts (
    record_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    display_name text NOT NULL,
    hostname text,
    aad_device_id text,
    fqdn text,
    entity_origin text DEFAULT 'entity_sheet'::text NOT NULL,
    seed_entity_mention_id uuid,
    host_state text NOT NULL,
    merged_into_record_id uuid,
    row_version bigint DEFAULT 1 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by_user_id uuid NOT NULL,
    updated_by_user_id uuid NOT NULL,
    location text,
    os_platform text,
    business_owner text,
    criticality text,
    containment_status text,
    CONSTRAINT hosts_entity_origin_core02_ck CHECK (entity_origin = ANY (ARRAY['entity_sheet'::text, 'entity_import'::text, 'created_from_mention'::text, 'system_upsert'::text])),
    CONSTRAINT hosts_host_state_check CHECK (host_state = ANY (ARRAY['stub'::text, 'canonical'::text, 'merged'::text])),
    CONSTRAINT hosts_merge_lineage_ck CHECK ((host_state = ANY (ARRAY['stub'::text, 'canonical'::text])) AND merged_into_record_id IS NULL OR host_state = 'merged'::text AND merged_into_record_id IS NOT NULL),
    CONSTRAINT hosts_pkey PRIMARY KEY (record_id),
    CONSTRAINT hosts_incident_identity_key UNIQUE (incident_id, record_id)
);

REVOKE ALL ON TABLE public.hosts FROM PUBLIC;

REVOKE ALL ON TYPE public.hosts FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.hosts TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.hosts TO cartulary_recovery;

CREATE TABLE public.identities (
    record_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    display_name text NOT NULL,
    upn text,
    email public.citext,
    sam_account_name text,
    aad_object_id text,
    sid text,
    entity_origin text DEFAULT 'entity_sheet'::text NOT NULL,
    seed_entity_mention_id uuid,
    identity_state text NOT NULL,
    merged_into_record_id uuid,
    row_version bigint DEFAULT 1 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by_user_id uuid NOT NULL,
    updated_by_user_id uuid NOT NULL,
    privilege_level text,
    mfa_state text,
    reset_status text,
    CONSTRAINT identities_entity_origin_core02_ck CHECK (entity_origin = ANY (ARRAY['entity_sheet'::text, 'entity_import'::text, 'created_from_mention'::text, 'system_upsert'::text])),
    CONSTRAINT identities_identity_state_check CHECK (identity_state = ANY (ARRAY['stub'::text, 'canonical'::text, 'merged'::text])),
    CONSTRAINT identities_merge_lineage_ck CHECK ((identity_state = ANY (ARRAY['stub'::text, 'canonical'::text])) AND merged_into_record_id IS NULL OR identity_state = 'merged'::text AND merged_into_record_id IS NOT NULL),
    CONSTRAINT identities_pkey PRIMARY KEY (record_id),
    CONSTRAINT identities_incident_identity_key UNIQUE (incident_id, record_id)
);

REVOKE ALL ON TABLE public.identities FROM PUBLIC;

REVOKE ALL ON TYPE public.identities FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.identities TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.identities TO cartulary_recovery;

-- +goose StatementBegin
CREATE FUNCTION public.entities_assert_entity_child_v1()
 RETURNS trigger
 LANGUAGE plpgsql
 SET search_path = pg_catalog, public
AS $$
DECLARE
    owner_valid boolean;
    source_incident_id uuid;
    resolved_incident_id uuid;
    resolved_record_type text;
    candidate text;
BEGIN
    IF TG_TABLE_NAME = 'entity_mentions' THEN
        SELECT incident_id
          INTO source_incident_id
          FROM public.records
         WHERE record_id = NEW.source_record_id;
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE = '23514', MESSAGE = 'entities_mention_source_invalid';
        END IF;
        IF NEW.resolution_status = 'resolved' THEN
            SELECT incident_id, record_type
              INTO resolved_incident_id, resolved_record_type
              FROM public.records
             WHERE record_id = NEW.resolved_record_id;
            IF NOT FOUND
                OR resolved_incident_id IS DISTINCT FROM source_incident_id
                OR resolved_record_type IS DISTINCT FROM NEW.entity_type THEN
                RAISE EXCEPTION USING ERRCODE = '23514', MESSAGE = 'entities_mention_target_invalid';
            END IF;
        END IF;
        RETURN NEW;
    END IF;

    IF NEW.entity_type = 'host' THEN
        SELECT EXISTS (
            SELECT 1
              FROM public.hosts h
             WHERE h.record_id = NEW.record_id
               AND h.incident_id = NEW.incident_id
        ) INTO owner_valid;
    ELSIF NEW.entity_type = 'identity' THEN
        SELECT EXISTS (
            SELECT 1
              FROM public.identities i
             WHERE i.record_id = NEW.record_id
               AND i.incident_id = NEW.incident_id
        ) INTO owner_valid;
    ELSE
        owner_valid := false;
    END IF;
    IF NOT owner_valid THEN
        RAISE EXCEPTION USING ERRCODE = '23514', MESSAGE = 'entities_child_owner_invalid';
    END IF;

    IF TG_TABLE_NAME = 'entity_aliases' THEN
        candidate := normalize(public.entities_trim_unicode_space_v1(NEW.raw_text), NFC);
        -- Leave the existing, field-specific empty/control/length CHECKs as the
        -- authoritative error source for values they already cover.  This
        -- trigger closes only the Unicode trim/NFC and exact-projection gap.
        IF candidate <> ''
            AND char_length(candidate) <= 256
            AND public.entities_source_codepoints_admitted_v1(candidate, false)
            AND (
                NEW.raw_text IS DISTINCT FROM candidate
                OR NEW.normalized_text::text IS DISTINCT FROM candidate
            ) THEN
            RAISE EXCEPTION USING ERRCODE = '23514', MESSAGE = 'entities_alias_normalization_invalid';
        END IF;
        RETURN NEW;
    END IF;

    IF NOT (
        (NEW.entity_type = 'host' AND NEW.identifier_type IN (
            'aad_device_id', 'fqdn', 'hostname'
        ))
        OR (NEW.entity_type = 'identity' AND NEW.identifier_type IN (
            'aad_object_id', 'sid', 'upn', 'email', 'sam_account_name'
        ))
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514', MESSAGE = 'entities_identifier_class_invalid';
    END IF;
    candidate := normalize(public.entities_trim_unicode_space_v1(NEW.raw_value), NFC);
    IF candidate = ''
        OR NOT public.entities_source_codepoints_admitted_v1(candidate, true) THEN
        RAISE EXCEPTION USING ERRCODE = '23514', MESSAGE = 'entities_identifier_value_invalid';
    END IF;
    -- Keep the existing invoker trigger and the normalizer's private ACL.
    -- Apply the same owner map here without granting a new helper capability.
    IF NEW.identifier_type = 'sid' THEN
        candidate := translate(candidate, 'abcdefghijklmnopqrstuvwxyzµßàáâãäåæçèéêëìíîïðñòóôõöøùúûüýþÿāăąćĉċčďđēĕėęěĝğġģĥħĩīĭįıĳĵķĺļľŀłńņňŋōŏőœŕŗřśŝşšţťŧũūŭůűųŵŷźżžſƀƃƅƈƌƒƕƙƚƛƞơƣƥƨƭưƴƶƹƽƿǅǆǈǉǋǌǎǐǒǔǖǘǚǜǝǟǡǣǥǧǩǫǭǯǲǳǵǹǻǽǿȁȃȅȇȉȋȍȏȑȓȕȗșțȝȟȣȥȧȩȫȭȯȱȳȼȿɀɂɇɉɋɍɏɐɑɒɓɔɖɗəɛɜɠɡɣɤɥɦɨɩɪɫɬɯɱɲɵɽʀʂʃʇʈʉʊʋʌʒʝʞͅͱͳͷͻͼͽάέήίαβγδεζηθικλμνξοπρςστυφχψωϊϋόύώϐϑϕϖϗϙϛϝϟϡϣϥϧϩϫϭϯϰϱϲϳϵϸϻабвгдежзийклмнопрстуфхцчшщъыьэюяѐёђѓєѕіїјљњћќѝўџѡѣѥѧѩѫѭѯѱѳѵѷѹѻѽѿҁҋҍҏґғҕҗҙқҝҟҡңҥҧҩҫҭүұҳҵҷҹһҽҿӂӄӆӈӊӌӎӏӑӓӕӗәӛӝӟӡӣӥӧөӫӭӯӱӳӵӷӹӻӽӿԁԃԅԇԉԋԍԏԑԓԕԗԙԛԝԟԡԣԥԧԩԫԭԯաբգդեզէըթժիլխծկհձղճմյնշոչպջռսվտրցւփքօֆაბგდევზთიკლმნოპჟრსტუფქღყშჩცძწჭხჯჰჱჲჳჴჵჶჷჸჹჺჽჾჿᏸᏹᏺᏻᏼᏽᲀᲁᲂᲃᲄᲅᲆᲇᲈᲊᵹᵽᶎḁḃḅḇḉḋḍḏḑḓḕḗḙḛḝḟḡḣḥḧḩḫḭḯḱḳḵḷḹḻḽḿṁṃṅṇṉṋṍṏṑṓṕṗṙṛṝṟṡṣṥṧṩṫṭṯṱṳṵṷṹṻṽṿẁẃẅẇẉẋẍẏẑẓẕẛạảấầẩẫậắằẳẵặẹẻẽếềểễệỉịọỏốồổỗộớờởỡợụủứừửữựỳỵỷỹỻỽỿἀἁἂἃἄἅἆἇἐἑἒἓἔἕἠἡἢἣἤἥἦἧἰἱἲἳἴἵἶἷὀὁὂὃὄὅὑὓὕὗὠὡὢὣὤὥὦὧὰάὲέὴήὶίὸόὺύὼώᾀᾁᾂᾃᾄᾅᾆᾇᾐᾑᾒᾓᾔᾕᾖᾗᾠᾡᾢᾣᾤᾥᾦᾧᾰᾱᾳιῃῐῑῠῡῥῳⅎⅰⅱⅲⅳⅴⅵⅶⅷⅸⅹⅺⅻⅼⅽⅾⅿↄⓐⓑⓒⓓⓔⓕⓖⓗⓘⓙⓚⓛⓜⓝⓞⓟⓠⓡⓢⓣⓤⓥⓦⓧⓨⓩⰰⰱⰲⰳⰴⰵⰶⰷⰸⰹⰺⰻⰼⰽⰾⰿⱀⱁⱂⱃⱄⱅⱆⱇⱈⱉⱊⱋⱌⱍⱎⱏⱐⱑⱒⱓⱔⱕⱖⱗⱘⱙⱚⱛⱜⱝⱞⱟⱡⱥⱦⱨⱪⱬⱳⱶⲁⲃⲅⲇⲉⲋⲍⲏⲑⲓⲕⲗⲙⲛⲝⲟⲡⲣⲥⲧⲩⲫⲭⲯⲱⲳⲵⲷⲹⲻⲽⲿⳁⳃⳅⳇⳉⳋⳍⳏⳑⳓⳕⳗⳙⳛⳝⳟⳡⳣⳬⳮⳳⴀⴁⴂⴃⴄⴅⴆⴇⴈⴉⴊⴋⴌⴍⴎⴏⴐⴑⴒⴓⴔⴕⴖⴗⴘⴙⴚⴛⴜⴝⴞⴟⴠⴡⴢⴣⴤⴥⴧⴭꙁꙃꙅꙇꙉꙋꙍꙏꙑꙓꙕꙗꙙꙛꙝꙟꙡꙣꙥꙧꙩꙫꙭꚁꚃꚅꚇꚉꚋꚍꚏꚑꚓꚕꚗꚙꚛꜣꜥꜧꜩꜫꜭꜯꜳꜵꜷꜹꜻꜽꜿꝁꝃꝅꝇꝉꝋꝍꝏꝑꝓꝕꝗꝙꝛꝝꝟꝡꝣꝥꝧꝩꝫꝭꝯꝺꝼꝿꞁꞃꞅꞇꞌꞑꞓꞔꞗꞙꞛꞝꞟꞡꞣꞥꞧꞩꞵꞷꞹꞻꞽꞿꟁꟃꟈꟊꟍꟑꟗꟙꟛꟶꭓꭰꭱꭲꭳꭴꭵꭶꭷꭸꭹꭺꭻꭼꭽꭾꭿꮀꮁꮂꮃꮄꮅꮆꮇꮈꮉꮊꮋꮌꮍꮎꮏꮐꮑꮒꮓꮔꮕꮖꮗꮘꮙꮚꮛꮜꮝꮞꮟꮠꮡꮢꮣꮤꮥꮦꮧꮨꮩꮪꮫꮬꮭꮮꮯꮰꮱꮲꮳꮴꮵꮶꮷꮸꮹꮺꮻꮼꮽꮾꮿａｂｃｄｅｆｇｈｉｊｋｌｍｎｏｐｑｒｓｔｕｖｗｘｙｚ𐐨𐐩𐐪𐐫𐐬𐐭𐐮𐐯𐐰𐐱𐐲𐐳𐐴𐐵𐐶𐐷𐐸𐐹𐐺𐐻𐐼𐐽𐐾𐐿𐑀𐑁𐑂𐑃𐑄𐑅𐑆𐑇𐑈𐑉𐑊𐑋𐑌𐑍𐑎𐑏𐓘𐓙𐓚𐓛𐓜𐓝𐓞𐓟𐓠𐓡𐓢𐓣𐓤𐓥𐓦𐓧𐓨𐓩𐓪𐓫𐓬𐓭𐓮𐓯𐓰𐓱𐓲𐓳𐓴𐓵𐓶𐓷𐓸𐓹𐓺𐓻𐖗𐖘𐖙𐖚𐖛𐖜𐖝𐖞𐖟𐖠𐖡𐖣𐖤𐖥𐖦𐖧𐖨𐖩𐖪𐖫𐖬𐖭𐖮𐖯𐖰𐖱𐖳𐖴𐖵𐖶𐖷𐖸𐖹𐖻𐖼𐳀𐳁𐳂𐳃𐳄𐳅𐳆𐳇𐳈𐳉𐳊𐳋𐳌𐳍𐳎𐳏𐳐𐳑𐳒𐳓𐳔𐳕𐳖𐳗𐳘𐳙𐳚𐳛𐳜𐳝𐳞𐳟𐳠𐳡𐳢𐳣𐳤𐳥𐳦𐳧𐳨𐳩𐳪𐳫𐳬𐳭𐳮𐳯𐳰𐳱𐳲𐵰𐵱𐵲𐵳𐵴𐵵𐵶𐵷𐵸𐵹𐵺𐵻𐵼𐵽𐵾𐵿𐶀𐶁𐶂𐶃𐶄𐶅𑣀𑣁𑣂𑣃𑣄𑣅𑣆𑣇𑣈𑣉𑣊𑣋𑣌𑣍𑣎𑣏𑣐𑣑𑣒𑣓𑣔𑣕𑣖𑣗𑣘𑣙𑣚𑣛𑣜𑣝𑣞𑣟𖹠𖹡𖹢𖹣𖹤𖹥𖹦𖹧𖹨𖹩𖹪𖹫𖹬𖹭𖹮𖹯𖹰𖹱𖹲𖹳𖹴𖹵𖹶𖹷𖹸𖹹𖹺𖹻𖹼𖹽𖹾𖹿𞤢𞤣𞤤𞤥𞤦𞤧𞤨𞤩𞤪𞤫𞤬𞤭𞤮𞤯𞤰𞤱𞤲𞤳𞤴𞤵𞤶𞤷𞤸𞤹𞤺𞤻𞤼𞤽𞤾𞤿𞥀𞥁𞥂𞥃', 'ABCDEFGHIJKLMNOPQRSTUVWXYZΜẞÀÁÂÃÄÅÆÇÈÉÊËÌÍÎÏÐÑÒÓÔÕÖØÙÚÛÜÝÞŸĀĂĄĆĈĊČĎĐĒĔĖĘĚĜĞĠĢĤĦĨĪĬĮIĲĴĶĹĻĽĿŁŃŅŇŊŌŎŐŒŔŖŘŚŜŞŠŢŤŦŨŪŬŮŰŲŴŶŹŻŽSɃƂƄƇƋƑǶƘȽꟜȠƠƢƤƧƬƯƳƵƸƼǷǄǄǇǇǊǊǍǏǑǓǕǗǙǛƎǞǠǢǤǦǨǪǬǮǱǱǴǸǺǼǾȀȂȄȆȈȊȌȎȐȒȔȖȘȚȜȞȢȤȦȨȪȬȮȰȲȻⱾⱿɁɆɈɊɌɎⱯⱭⱰƁƆƉƊƏƐꞫƓꞬƔꟋꞍꞪƗƖꞮⱢꞭƜⱮƝƟⱤƦꟅƩꞱƮɄƱƲɅƷꞲꞰΙͰͲͶϽϾϿΆΈΉΊΑΒΓΔΕΖΗΘΙΚΛΜΝΞΟΠΡΣΣΤΥΦΧΨΩΪΫΌΎΏΒΘΦΠϏϘϚϜϞϠϢϤϦϨϪϬϮΚΡϹͿΕϷϺАБВГДЕЖЗИЙКЛМНОПРСТУФХЦЧШЩЪЫЬЭЮЯЀЁЂЃЄЅІЇЈЉЊЋЌЍЎЏѠѢѤѦѨѪѬѮѰѲѴѶѸѺѼѾҀҊҌҎҐҒҔҖҘҚҜҞҠҢҤҦҨҪҬҮҰҲҴҶҸҺҼҾӁӃӅӇӉӋӍӀӐӒӔӖӘӚӜӞӠӢӤӦӨӪӬӮӰӲӴӶӸӺӼӾԀԂԄԆԈԊԌԎԐԒԔԖԘԚԜԞԠԢԤԦԨԪԬԮԱԲԳԴԵԶԷԸԹԺԻԼԽԾԿՀՁՂՃՄՅՆՇՈՉՊՋՌՍՎՏՐՑՒՓՔՕՖᲐᲑᲒᲓᲔᲕᲖᲗᲘᲙᲚᲛᲜᲝᲞᲟᲠᲡᲢᲣᲤᲥᲦᲧᲨᲩᲪᲫᲬᲭᲮᲯᲰᲱᲲᲳᲴᲵᲶᲷᲸᲹᲺᲽᲾᲿᏰᏱᏲᏳᏴᏵВДОСТТЪѢꙊᲉꝽⱣꟆḀḂḄḆḈḊḌḎḐḒḔḖḘḚḜḞḠḢḤḦḨḪḬḮḰḲḴḶḸḺḼḾṀṂṄṆṈṊṌṎṐṒṔṖṘṚṜṞṠṢṤṦṨṪṬṮṰṲṴṶṸṺṼṾẀẂẄẆẈẊẌẎẐẒẔṠẠẢẤẦẨẪẬẮẰẲẴẶẸẺẼẾỀỂỄỆỈỊỌỎỐỒỔỖỘỚỜỞỠỢỤỦỨỪỬỮỰỲỴỶỸỺỼỾἈἉἊἋἌἍἎἏἘἙἚἛἜἝἨἩἪἫἬἭἮἯἸἹἺἻἼἽἾἿὈὉὊὋὌὍὙὛὝὟὨὩὪὫὬὭὮὯᾺΆῈΈῊΉῚΊῸΌῪΎῺΏᾈᾉᾊᾋᾌᾍᾎᾏᾘᾙᾚᾛᾜᾝᾞᾟᾨᾩᾪᾫᾬᾭᾮᾯᾸᾹᾼΙῌῘῙῨῩῬῼℲⅠⅡⅢⅣⅤⅥⅦⅧⅨⅩⅪⅫⅬⅭⅮⅯↃⒶⒷⒸⒹⒺⒻⒼⒽⒾⒿⓀⓁⓂⓃⓄⓅⓆⓇⓈⓉⓊⓋⓌⓍⓎⓏⰀⰁⰂⰃⰄⰅⰆⰇⰈⰉⰊⰋⰌⰍⰎⰏⰐⰑⰒⰓⰔⰕⰖⰗⰘⰙⰚⰛⰜⰝⰞⰟⰠⰡⰢⰣⰤⰥⰦⰧⰨⰩⰪⰫⰬⰭⰮⰯⱠȺȾⱧⱩⱫⱲⱵⲀⲂⲄⲆⲈⲊⲌⲎⲐⲒⲔⲖⲘⲚⲜⲞⲠⲢⲤⲦⲨⲪⲬⲮⲰⲲⲴⲶⲸⲺⲼⲾⳀⳂⳄⳆⳈⳊⳌⳎⳐⳒⳔⳖⳘⳚⳜⳞⳠⳢⳫⳭⳲႠႡႢႣႤႥႦႧႨႩႪႫႬႭႮႯႰႱႲႳႴႵႶႷႸႹႺႻႼႽႾႿჀჁჂჃჄჅჇჍꙀꙂꙄꙆꙈꙊꙌꙎꙐꙒꙔꙖꙘꙚꙜꙞꙠꙢꙤꙦꙨꙪꙬꚀꚂꚄꚆꚈꚊꚌꚎꚐꚒꚔꚖꚘꚚꜢꜤꜦꜨꜪꜬꜮꜲꜴꜶꜸꜺꜼꜾꝀꝂꝄꝆꝈꝊꝌꝎꝐꝒꝔꝖꝘꝚꝜꝞꝠꝢꝤꝦꝨꝪꝬꝮꝹꝻꝾꞀꞂꞄꞆꞋꞐꞒꟄꞖꞘꞚꞜꞞꞠꞢꞤꞦꞨꞴꞶꞸꞺꞼꞾꟀꟂꟇꟉꟌꟐꟖꟘꟚꟵꞳᎠᎡᎢᎣᎤᎥᎦᎧᎨᎩᎪᎫᎬᎭᎮᎯᎰᎱᎲᎳᎴᎵᎶᎷᎸᎹᎺᎻᎼᎽᎾᎿᏀᏁᏂᏃᏄᏅᏆᏇᏈᏉᏊᏋᏌᏍᏎᏏᏐᏑᏒᏓᏔᏕᏖᏗᏘᏙᏚᏛᏜᏝᏞᏟᏠᏡᏢᏣᏤᏥᏦᏧᏨᏩᏪᏫᏬᏭᏮᏯＡＢＣＤＥＦＧＨＩＪＫＬＭＮＯＰＱＲＳＴＵＶＷＸＹＺ𐐀𐐁𐐂𐐃𐐄𐐅𐐆𐐇𐐈𐐉𐐊𐐋𐐌𐐍𐐎𐐏𐐐𐐑𐐒𐐓𐐔𐐕𐐖𐐗𐐘𐐙𐐚𐐛𐐜𐐝𐐞𐐟𐐠𐐡𐐢𐐣𐐤𐐥𐐦𐐧𐒰𐒱𐒲𐒳𐒴𐒵𐒶𐒷𐒸𐒹𐒺𐒻𐒼𐒽𐒾𐒿𐓀𐓁𐓂𐓃𐓄𐓅𐓆𐓇𐓈𐓉𐓊𐓋𐓌𐓍𐓎𐓏𐓐𐓑𐓒𐓓𐕰𐕱𐕲𐕳𐕴𐕵𐕶𐕷𐕸𐕹𐕺𐕼𐕽𐕾𐕿𐖀𐖁𐖂𐖃𐖄𐖅𐖆𐖇𐖈𐖉𐖊𐖌𐖍𐖎𐖏𐖐𐖑𐖒𐖔𐖕𐲀𐲁𐲂𐲃𐲄𐲅𐲆𐲇𐲈𐲉𐲊𐲋𐲌𐲍𐲎𐲏𐲐𐲑𐲒𐲓𐲔𐲕𐲖𐲗𐲘𐲙𐲚𐲛𐲜𐲝𐲞𐲟𐲠𐲡𐲢𐲣𐲤𐲥𐲦𐲧𐲨𐲩𐲪𐲫𐲬𐲭𐲮𐲯𐲰𐲱𐲲𐵐𐵑𐵒𐵓𐵔𐵕𐵖𐵗𐵘𐵙𐵚𐵛𐵜𐵝𐵞𐵟𐵠𐵡𐵢𐵣𐵤𐵥𑢠𑢡𑢢𑢣𑢤𑢥𑢦𑢧𑢨𑢩𑢪𑢫𑢬𑢭𑢮𑢯𑢰𑢱𑢲𑢳𑢴𑢵𑢶𑢷𑢸𑢹𑢺𑢻𑢼𑢽𑢾𑢿𖹀𖹁𖹂𖹃𖹄𖹅𖹆𖹇𖹈𖹉𖹊𖹋𖹌𖹍𖹎𖹏𖹐𖹑𖹒𖹓𖹔𖹕𖹖𖹗𖹘𖹙𖹚𖹛𖹜𖹝𖹞𖹟𞤀𞤁𞤂𞤃𞤄𞤅𞤆𞤇𞤈𞤉𞤊𞤋𞤌𞤍𞤎𞤏𞤐𞤑𞤒𞤓𞤔𞤕𞤖𞤗𞤘𞤙𞤚𞤛𞤜𞤝𞤞𞤟𞤠𞤡');
    ELSE
        candidate := translate(candidate, 'ABCDEFGHIJKLMNOPQRSTUVWXYZÀÁÂÃÄÅÆÇÈÉÊËÌÍÎÏÐÑÒÓÔÕÖØÙÚÛÜÝÞĀĂĄĆĈĊČĎĐĒĔĖĘĚĜĞĠĢĤĦĨĪĬĮİĲĴĶĹĻĽĿŁŃŅŇŊŌŎŐŒŔŖŘŚŜŞŠŢŤŦŨŪŬŮŰŲŴŶŸŹŻŽƁƂƄƆƇƉƊƋƎƏƐƑƓƔƖƗƘƜƝƟƠƢƤƦƧƩƬƮƯƱƲƳƵƷƸƼǄǅǇǈǊǋǍǏǑǓǕǗǙǛǞǠǢǤǦǨǪǬǮǱǲǴǶǷǸǺǼǾȀȂȄȆȈȊȌȎȐȒȔȖȘȚȜȞȠȢȤȦȨȪȬȮȰȲȺȻȽȾɁɃɄɅɆɈɊɌɎͰͲͶͿΆΈΉΊΌΎΏΑΒΓΔΕΖΗΘΙΚΛΜΝΞΟΠΡΣΤΥΦΧΨΩΪΫϏϘϚϜϞϠϢϤϦϨϪϬϮϴϷϹϺϽϾϿЀЁЂЃЄЅІЇЈЉЊЋЌЍЎЏАБВГДЕЖЗИЙКЛМНОПРСТУФХЦЧШЩЪЫЬЭЮЯѠѢѤѦѨѪѬѮѰѲѴѶѸѺѼѾҀҊҌҎҐҒҔҖҘҚҜҞҠҢҤҦҨҪҬҮҰҲҴҶҸҺҼҾӀӁӃӅӇӉӋӍӐӒӔӖӘӚӜӞӠӢӤӦӨӪӬӮӰӲӴӶӸӺӼӾԀԂԄԆԈԊԌԎԐԒԔԖԘԚԜԞԠԢԤԦԨԪԬԮԱԲԳԴԵԶԷԸԹԺԻԼԽԾԿՀՁՂՃՄՅՆՇՈՉՊՋՌՍՎՏՐՑՒՓՔՕՖႠႡႢႣႤႥႦႧႨႩႪႫႬႭႮႯႰႱႲႳႴႵႶႷႸႹႺႻႼႽႾႿჀჁჂჃჄჅჇჍᎠᎡᎢᎣᎤᎥᎦᎧᎨᎩᎪᎫᎬᎭᎮᎯᎰᎱᎲᎳᎴᎵᎶᎷᎸᎹᎺᎻᎼᎽᎾᎿᏀᏁᏂᏃᏄᏅᏆᏇᏈᏉᏊᏋᏌᏍᏎᏏᏐᏑᏒᏓᏔᏕᏖᏗᏘᏙᏚᏛᏜᏝᏞᏟᏠᏡᏢᏣᏤᏥᏦᏧᏨᏩᏪᏫᏬᏭᏮᏯᏰᏱᏲᏳᏴᏵᲉᲐᲑᲒᲓᲔᲕᲖᲗᲘᲙᲚᲛᲜᲝᲞᲟᲠᲡᲢᲣᲤᲥᲦᲧᲨᲩᲪᲫᲬᲭᲮᲯᲰᲱᲲᲳᲴᲵᲶᲷᲸᲹᲺᲽᲾᲿḀḂḄḆḈḊḌḎḐḒḔḖḘḚḜḞḠḢḤḦḨḪḬḮḰḲḴḶḸḺḼḾṀṂṄṆṈṊṌṎṐṒṔṖṘṚṜṞṠṢṤṦṨṪṬṮṰṲṴṶṸṺṼṾẀẂẄẆẈẊẌẎẐẒẔẞẠẢẤẦẨẪẬẮẰẲẴẶẸẺẼẾỀỂỄỆỈỊỌỎỐỒỔỖỘỚỜỞỠỢỤỦỨỪỬỮỰỲỴỶỸỺỼỾἈἉἊἋἌἍἎἏἘἙἚἛἜἝἨἩἪἫἬἭἮἯἸἹἺἻἼἽἾἿὈὉὊὋὌὍὙὛὝὟὨὩὪὫὬὭὮὯᾈᾉᾊᾋᾌᾍᾎᾏᾘᾙᾚᾛᾜᾝᾞᾟᾨᾩᾪᾫᾬᾭᾮᾯᾸᾹᾺΆᾼῈΈῊΉῌῘῙῚΊῨῩῪΎῬῸΌῺΏῼΩKÅℲⅠⅡⅢⅣⅤⅥⅦⅧⅨⅩⅪⅫⅬⅭⅮⅯↃⒶⒷⒸⒹⒺⒻⒼⒽⒾⒿⓀⓁⓂⓃⓄⓅⓆⓇⓈⓉⓊⓋⓌⓍⓎⓏⰀⰁⰂⰃⰄⰅⰆⰇⰈⰉⰊⰋⰌⰍⰎⰏⰐⰑⰒⰓⰔⰕⰖⰗⰘⰙⰚⰛⰜⰝⰞⰟⰠⰡⰢⰣⰤⰥⰦⰧⰨⰩⰪⰫⰬⰭⰮⰯⱠⱢⱣⱤⱧⱩⱫⱭⱮⱯⱰⱲⱵⱾⱿⲀⲂⲄⲆⲈⲊⲌⲎⲐⲒⲔⲖⲘⲚⲜⲞⲠⲢⲤⲦⲨⲪⲬⲮⲰⲲⲴⲶⲸⲺⲼⲾⳀⳂⳄⳆⳈⳊⳌⳎⳐⳒⳔⳖⳘⳚⳜⳞⳠⳢⳫⳭⳲꙀꙂꙄꙆꙈꙊꙌꙎꙐꙒꙔꙖꙘꙚꙜꙞꙠꙢꙤꙦꙨꙪꙬꚀꚂꚄꚆꚈꚊꚌꚎꚐꚒꚔꚖꚘꚚꜢꜤꜦꜨꜪꜬꜮꜲꜴꜶꜸꜺꜼꜾꝀꝂꝄꝆꝈꝊꝌꝎꝐꝒꝔꝖꝘꝚꝜꝞꝠꝢꝤꝦꝨꝪꝬꝮꝹꝻꝽꝾꞀꞂꞄꞆꞋꞍꞐꞒꞖꞘꞚꞜꞞꞠꞢꞤꞦꞨꞪꞫꞬꞭꞮꞰꞱꞲꞳꞴꞶꞸꞺꞼꞾꟀꟂꟄꟅꟆꟇꟉꟋꟌꟐꟖꟘꟚꟜꟵＡＢＣＤＥＦＧＨＩＪＫＬＭＮＯＰＱＲＳＴＵＶＷＸＹＺ𐐀𐐁𐐂𐐃𐐄𐐅𐐆𐐇𐐈𐐉𐐊𐐋𐐌𐐍𐐎𐐏𐐐𐐑𐐒𐐓𐐔𐐕𐐖𐐗𐐘𐐙𐐚𐐛𐐜𐐝𐐞𐐟𐐠𐐡𐐢𐐣𐐤𐐥𐐦𐐧𐒰𐒱𐒲𐒳𐒴𐒵𐒶𐒷𐒸𐒹𐒺𐒻𐒼𐒽𐒾𐒿𐓀𐓁𐓂𐓃𐓄𐓅𐓆𐓇𐓈𐓉𐓊𐓋𐓌𐓍𐓎𐓏𐓐𐓑𐓒𐓓𐕰𐕱𐕲𐕳𐕴𐕵𐕶𐕷𐕸𐕹𐕺𐕼𐕽𐕾𐕿𐖀𐖁𐖂𐖃𐖄𐖅𐖆𐖇𐖈𐖉𐖊𐖌𐖍𐖎𐖏𐖐𐖑𐖒𐖔𐖕𐲀𐲁𐲂𐲃𐲄𐲅𐲆𐲇𐲈𐲉𐲊𐲋𐲌𐲍𐲎𐲏𐲐𐲑𐲒𐲓𐲔𐲕𐲖𐲗𐲘𐲙𐲚𐲛𐲜𐲝𐲞𐲟𐲠𐲡𐲢𐲣𐲤𐲥𐲦𐲧𐲨𐲩𐲪𐲫𐲬𐲭𐲮𐲯𐲰𐲱𐲲𐵐𐵑𐵒𐵓𐵔𐵕𐵖𐵗𐵘𐵙𐵚𐵛𐵜𐵝𐵞𐵟𐵠𐵡𐵢𐵣𐵤𐵥𑢠𑢡𑢢𑢣𑢤𑢥𑢦𑢧𑢨𑢩𑢪𑢫𑢬𑢭𑢮𑢯𑢰𑢱𑢲𑢳𑢴𑢵𑢶𑢷𑢸𑢹𑢺𑢻𑢼𑢽𑢾𑢿𖹀𖹁𖹂𖹃𖹄𖹅𖹆𖹇𖹈𖹉𖹊𖹋𖹌𖹍𖹎𖹏𖹐𖹑𖹒𖹓𖹔𖹕𖹖𖹗𖹘𖹙𖹚𖹛𖹜𖹝𖹞𖹟𞤀𞤁𞤂𞤃𞤄𞤅𞤆𞤇𞤈𞤉𞤊𞤋𞤌𞤍𞤎𞤏𞤐𞤑𞤒𞤓𞤔𞤕𞤖𞤗𞤘𞤙𞤚𞤛𞤜𞤝𞤞𞤟𞤠𞤡', 'abcdefghijklmnopqrstuvwxyzàáâãäåæçèéêëìíîïðñòóôõöøùúûüýþāăąćĉċčďđēĕėęěĝğġģĥħĩīĭįiĳĵķĺļľŀłńņňŋōŏőœŕŗřśŝşšţťŧũūŭůűųŵŷÿźżžɓƃƅɔƈɖɗƌǝəɛƒɠɣɩɨƙɯɲɵơƣƥʀƨʃƭʈưʊʋƴƶʒƹƽǆǆǉǉǌǌǎǐǒǔǖǘǚǜǟǡǣǥǧǩǫǭǯǳǳǵƕƿǹǻǽǿȁȃȅȇȉȋȍȏȑȓȕȗșțȝȟƞȣȥȧȩȫȭȯȱȳⱥȼƚⱦɂƀʉʌɇɉɋɍɏͱͳͷϳάέήίόύώαβγδεζηθικλμνξοπρστυφχψωϊϋϗϙϛϝϟϡϣϥϧϩϫϭϯθϸϲϻͻͼͽѐёђѓєѕіїјљњћќѝўџабвгдежзийклмнопрстуфхцчшщъыьэюяѡѣѥѧѩѫѭѯѱѳѵѷѹѻѽѿҁҋҍҏґғҕҗҙқҝҟҡңҥҧҩҫҭүұҳҵҷҹһҽҿӏӂӄӆӈӊӌӎӑӓӕӗәӛӝӟӡӣӥӧөӫӭӯӱӳӵӷӹӻӽӿԁԃԅԇԉԋԍԏԑԓԕԗԙԛԝԟԡԣԥԧԩԫԭԯաբգդեզէըթժիլխծկհձղճմյնշոչպջռսվտրցւփքօֆⴀⴁⴂⴃⴄⴅⴆⴇⴈⴉⴊⴋⴌⴍⴎⴏⴐⴑⴒⴓⴔⴕⴖⴗⴘⴙⴚⴛⴜⴝⴞⴟⴠⴡⴢⴣⴤⴥⴧⴭꭰꭱꭲꭳꭴꭵꭶꭷꭸꭹꭺꭻꭼꭽꭾꭿꮀꮁꮂꮃꮄꮅꮆꮇꮈꮉꮊꮋꮌꮍꮎꮏꮐꮑꮒꮓꮔꮕꮖꮗꮘꮙꮚꮛꮜꮝꮞꮟꮠꮡꮢꮣꮤꮥꮦꮧꮨꮩꮪꮫꮬꮭꮮꮯꮰꮱꮲꮳꮴꮵꮶꮷꮸꮹꮺꮻꮼꮽꮾꮿᏸᏹᏺᏻᏼᏽᲊაბგდევზთიკლმნოპჟრსტუფქღყშჩცძწჭხჯჰჱჲჳჴჵჶჷჸჹჺჽჾჿḁḃḅḇḉḋḍḏḑḓḕḗḙḛḝḟḡḣḥḧḩḫḭḯḱḳḵḷḹḻḽḿṁṃṅṇṉṋṍṏṑṓṕṗṙṛṝṟṡṣṥṧṩṫṭṯṱṳṵṷṹṻṽṿẁẃẅẇẉẋẍẏẑẓẕßạảấầẩẫậắằẳẵặẹẻẽếềểễệỉịọỏốồổỗộớờởỡợụủứừửữựỳỵỷỹỻỽỿἀἁἂἃἄἅἆἇἐἑἒἓἔἕἠἡἢἣἤἥἦἧἰἱἲἳἴἵἶἷὀὁὂὃὄὅὑὓὕὗὠὡὢὣὤὥὦὧᾀᾁᾂᾃᾄᾅᾆᾇᾐᾑᾒᾓᾔᾕᾖᾗᾠᾡᾢᾣᾤᾥᾦᾧᾰᾱὰάᾳὲέὴήῃῐῑὶίῠῡὺύῥὸόὼώῳωkåⅎⅰⅱⅲⅳⅴⅵⅶⅷⅸⅹⅺⅻⅼⅽⅾⅿↄⓐⓑⓒⓓⓔⓕⓖⓗⓘⓙⓚⓛⓜⓝⓞⓟⓠⓡⓢⓣⓤⓥⓦⓧⓨⓩⰰⰱⰲⰳⰴⰵⰶⰷⰸⰹⰺⰻⰼⰽⰾⰿⱀⱁⱂⱃⱄⱅⱆⱇⱈⱉⱊⱋⱌⱍⱎⱏⱐⱑⱒⱓⱔⱕⱖⱗⱘⱙⱚⱛⱜⱝⱞⱟⱡɫᵽɽⱨⱪⱬɑɱɐɒⱳⱶȿɀⲁⲃⲅⲇⲉⲋⲍⲏⲑⲓⲕⲗⲙⲛⲝⲟⲡⲣⲥⲧⲩⲫⲭⲯⲱⲳⲵⲷⲹⲻⲽⲿⳁⳃⳅⳇⳉⳋⳍⳏⳑⳓⳕⳗⳙⳛⳝⳟⳡⳣⳬⳮⳳꙁꙃꙅꙇꙉꙋꙍꙏꙑꙓꙕꙗꙙꙛꙝꙟꙡꙣꙥꙧꙩꙫꙭꚁꚃꚅꚇꚉꚋꚍꚏꚑꚓꚕꚗꚙꚛꜣꜥꜧꜩꜫꜭꜯꜳꜵꜷꜹꜻꜽꜿꝁꝃꝅꝇꝉꝋꝍꝏꝑꝓꝕꝗꝙꝛꝝꝟꝡꝣꝥꝧꝩꝫꝭꝯꝺꝼᵹꝿꞁꞃꞅꞇꞌɥꞑꞓꞗꞙꞛꞝꞟꞡꞣꞥꞧꞩɦɜɡɬɪʞʇʝꭓꞵꞷꞹꞻꞽꞿꟁꟃꞔʂᶎꟈꟊɤꟍꟑꟗꟙꟛƛꟶａｂｃｄｅｆｇｈｉｊｋｌｍｎｏｐｑｒｓｔｕｖｗｘｙｚ𐐨𐐩𐐪𐐫𐐬𐐭𐐮𐐯𐐰𐐱𐐲𐐳𐐴𐐵𐐶𐐷𐐸𐐹𐐺𐐻𐐼𐐽𐐾𐐿𐑀𐑁𐑂𐑃𐑄𐑅𐑆𐑇𐑈𐑉𐑊𐑋𐑌𐑍𐑎𐑏𐓘𐓙𐓚𐓛𐓜𐓝𐓞𐓟𐓠𐓡𐓢𐓣𐓤𐓥𐓦𐓧𐓨𐓩𐓪𐓫𐓬𐓭𐓮𐓯𐓰𐓱𐓲𐓳𐓴𐓵𐓶𐓷𐓸𐓹𐓺𐓻𐖗𐖘𐖙𐖚𐖛𐖜𐖝𐖞𐖟𐖠𐖡𐖣𐖤𐖥𐖦𐖧𐖨𐖩𐖪𐖫𐖬𐖭𐖮𐖯𐖰𐖱𐖳𐖴𐖵𐖶𐖷𐖸𐖹𐖻𐖼𐳀𐳁𐳂𐳃𐳄𐳅𐳆𐳇𐳈𐳉𐳊𐳋𐳌𐳍𐳎𐳏𐳐𐳑𐳒𐳓𐳔𐳕𐳖𐳗𐳘𐳙𐳚𐳛𐳜𐳝𐳞𐳟𐳠𐳡𐳢𐳣𐳤𐳥𐳦𐳧𐳨𐳩𐳪𐳫𐳬𐳭𐳮𐳯𐳰𐳱𐳲𐵰𐵱𐵲𐵳𐵴𐵵𐵶𐵷𐵸𐵹𐵺𐵻𐵼𐵽𐵾𐵿𐶀𐶁𐶂𐶃𐶄𐶅𑣀𑣁𑣂𑣃𑣄𑣅𑣆𑣇𑣈𑣉𑣊𑣋𑣌𑣍𑣎𑣏𑣐𑣑𑣒𑣓𑣔𑣕𑣖𑣗𑣘𑣙𑣚𑣛𑣜𑣝𑣞𑣟𖹠𖹡𖹢𖹣𖹤𖹥𖹦𖹧𖹨𖹩𖹪𖹫𖹬𖹭𖹮𖹯𖹰𖹱𖹲𖹳𖹴𖹵𖹶𖹷𖹸𖹹𖹺𖹻𖹼𖹽𖹾𖹿𞤢𞤣𞤤𞤥𞤦𞤧𞤨𞤩𞤪𞤫𞤬𞤭𞤮𞤯𞤰𞤱𞤲𞤳𞤴𞤵𞤶𞤷𞤸𞤹𞤺𞤻𞤼𞤽𞤾𞤿𞥀𞥁𞥂𞥃');
    END IF;
    IF NEW.normalized_value IS DISTINCT FROM candidate THEN
        RAISE EXCEPTION USING ERRCODE = '23514', MESSAGE = 'entities_identifier_normalization_invalid';
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.entities_assert_entity_child_v1() FROM PUBLIC;

-- +goose StatementBegin
CREATE FUNCTION public.entities_assert_entity_envelope_update_v1()
 RETURNS trigger
 LANGUAGE plpgsql
 SET search_path = pg_catalog, public
AS $$
DECLARE
    source_count integer;
BEGIN
    IF NEW.record_type = 'host' THEN
        SELECT count(*)
          INTO source_count
          FROM public.hosts h
         WHERE h.record_id = NEW.record_id
           AND h.incident_id = NEW.incident_id
           AND h.row_version = NEW.row_version
           AND h.created_at = NEW.created_at
           AND h.updated_at = NEW.updated_at
           AND h.created_by_user_id = NEW.created_by_user_id
           AND h.updated_by_user_id = NEW.updated_by_user_id;
    ELSIF NEW.record_type = 'identity' THEN
        SELECT count(*)
          INTO source_count
          FROM public.identities i
         WHERE i.record_id = NEW.record_id
           AND i.incident_id = NEW.incident_id
           AND i.row_version = NEW.row_version
           AND i.created_at = NEW.created_at
           AND i.updated_at = NEW.updated_at
           AND i.created_by_user_id = NEW.created_by_user_id
           AND i.updated_by_user_id = NEW.updated_by_user_id;
    ELSE
        RETURN NEW;
    END IF;
    IF source_count <> 1 THEN
        RAISE EXCEPTION USING
            ERRCODE = '23514',
            MESSAGE = 'entities_envelope_source_mismatch';
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.entities_assert_entity_envelope_update_v1() FROM PUBLIC;

-- +goose StatementBegin
CREATE FUNCTION public.entities_assert_entity_source_v1()
 RETURNS trigger
 LANGUAGE plpgsql
 SET search_path = pg_catalog, public
AS $$
DECLARE
    envelope public.records%ROWTYPE;
    target_incident_id uuid;
    target_state text;
BEGIN
    SELECT *
      INTO envelope
      FROM public.records
     WHERE record_id = NEW.record_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING
            ERRCODE = '23514',
            MESSAGE = 'entities_source_envelope_missing';
    END IF;
    IF envelope.incident_id IS DISTINCT FROM NEW.incident_id
        OR envelope.record_type IS DISTINCT FROM TG_ARGV[0]
        OR envelope.row_version IS DISTINCT FROM NEW.row_version
        OR envelope.created_at IS DISTINCT FROM NEW.created_at
        OR envelope.updated_at IS DISTINCT FROM NEW.updated_at
        OR envelope.created_by_user_id IS DISTINCT FROM NEW.created_by_user_id
        OR envelope.updated_by_user_id IS DISTINCT FROM NEW.updated_by_user_id THEN
        RAISE EXCEPTION USING
            ERRCODE = '23514',
            MESSAGE = 'entities_source_envelope_mismatch';
    END IF;

    IF TG_ARGV[0] = 'host' THEN
        IF NEW.host_state = 'merged' THEN
            IF NEW.merged_into_record_id = NEW.record_id THEN
                RAISE EXCEPTION USING ERRCODE = '23514', MESSAGE = 'entities_source_merge_self_reference';
            END IF;
            SELECT incident_id, host_state
              INTO target_incident_id, target_state
              FROM public.hosts
             WHERE record_id = NEW.merged_into_record_id;
        ELSE
            RETURN NEW;
        END IF;
    ELSE
        IF NEW.identity_state = 'merged' THEN
            IF NEW.merged_into_record_id = NEW.record_id THEN
                RAISE EXCEPTION USING ERRCODE = '23514', MESSAGE = 'entities_source_merge_self_reference';
            END IF;
            SELECT incident_id, identity_state
              INTO target_incident_id, target_state
              FROM public.identities
             WHERE record_id = NEW.merged_into_record_id;
        ELSE
            RETURN NEW;
        END IF;
    END IF;
    IF NOT FOUND
        OR target_incident_id IS DISTINCT FROM NEW.incident_id
        OR target_state = 'merged' THEN
        RAISE EXCEPTION USING
            ERRCODE = '23514',
            MESSAGE = 'entities_source_merge_target_invalid';
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.entities_assert_entity_source_v1() FROM PUBLIC;

-- +goose StatementBegin
CREATE FUNCTION public.entities_expected_active_identifier_claims_v1(target_record_id uuid)
 RETURNS TABLE(incident_id uuid, entity_type text, identifier_type text, normalized_value text, record_id uuid)
 LANGUAGE sql
 STABLE
 SET search_path = pg_catalog, public
AS $$
    SELECT DISTINCT expected.incident_id, expected.entity_type,
           expected.identifier_type, expected.normalized_value,
           expected.record_id
      FROM (
        SELECT h.incident_id, 'host'::text AS entity_type,
               candidate.identifier_type,
               public.entities_normalize_identifier_v1(
                   candidate.identifier_type,
                   candidate.raw_value
               ) AS normalized_value,
               h.record_id
          FROM public.hosts h
          JOIN public.records r
            ON r.record_id = h.record_id
           AND r.incident_id = h.incident_id
           AND r.record_type = 'host'
          CROSS JOIN LATERAL (VALUES
              ('aad_device_id'::text, h.aad_device_id),
              ('fqdn'::text, h.fqdn),
              ('hostname'::text, h.hostname)
          ) AS candidate(identifier_type, raw_value)
         WHERE r.deleted_at IS NULL
           AND h.host_state IN ('stub', 'canonical')
           AND candidate.raw_value IS NOT NULL
           AND (target_record_id IS NULL OR h.record_id = target_record_id)
        UNION ALL
        SELECT i.incident_id, 'identity'::text,
               candidate.identifier_type,
               public.entities_normalize_identifier_v1(
                   candidate.identifier_type,
                   candidate.raw_value
               ),
               i.record_id
          FROM public.identities i
          JOIN public.records r
            ON r.record_id = i.record_id
           AND r.incident_id = i.incident_id
           AND r.record_type = 'identity'
          CROSS JOIN LATERAL (VALUES
              ('aad_object_id'::text, i.aad_object_id),
              ('sid'::text, i.sid),
              ('upn'::text, i.upn),
              ('email'::text, i.email::text),
              ('sam_account_name'::text, i.sam_account_name)
          ) AS candidate(identifier_type, raw_value)
         WHERE r.deleted_at IS NULL
           AND i.identity_state IN ('stub', 'canonical')
           AND candidate.raw_value IS NOT NULL
           AND (target_record_id IS NULL OR i.record_id = target_record_id)
        UNION ALL
        SELECT p.incident_id, p.entity_type, p.identifier_type,
               p.normalized_value, p.record_id
          FROM public.entity_preserved_identifiers p
          JOIN public.records r
            ON r.record_id = p.record_id
           AND r.incident_id = p.incident_id
           AND r.record_type = p.entity_type
          LEFT JOIN public.hosts h
            ON p.entity_type = 'host'
           AND h.record_id = p.record_id
          LEFT JOIN public.identities i
            ON p.entity_type = 'identity'
           AND i.record_id = p.record_id
         WHERE r.deleted_at IS NULL
           AND p.deleted_at IS NULL
           AND p.classification = 'exact_match_reuse'
           AND (target_record_id IS NULL OR p.record_id = target_record_id)
           AND (
               (p.entity_type = 'host'
                   AND h.host_state IN ('stub', 'canonical'))
               OR (p.entity_type = 'identity'
                   AND i.identity_state IN ('stub', 'canonical'))
           )
      ) AS expected
     WHERE expected.normalized_value IS NOT NULL
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.entities_expected_active_identifier_claims_v1(target_record_id uuid) FROM PUBLIC;

-- +goose StatementBegin
CREATE FUNCTION public.entities_reject_entity_source_delete_v1()
 RETURNS trigger
 LANGUAGE plpgsql
 SET search_path = pg_catalog, public
AS $$
BEGIN
    IF EXISTS (
        SELECT 1
          FROM public.records r
         WHERE r.record_id = OLD.record_id
           AND r.record_type = TG_ARGV[0]
    ) THEN
        RAISE EXCEPTION USING
            ERRCODE = '23514',
            MESSAGE = 'entities_source_delete_requires_envelope_delete';
    END IF;
    RETURN OLD;
END;
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.entities_reject_entity_source_delete_v1() FROM PUBLIC;

-- +goose StatementBegin
CREATE FUNCTION public.entities_release_active_identifier_claims_v1(target_record_id uuid)
 RETURNS bigint
 LANGUAGE plpgsql
 SECURITY DEFINER
 SET search_path = pg_catalog, public
AS $$
DECLARE
    released_count bigint;
BEGIN
    DELETE FROM public.entity_active_identifier_claims
     WHERE record_id = target_record_id;
    GET DIAGNOSTICS released_count = ROW_COUNT;
    RETURN released_count;
END;
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.entities_release_active_identifier_claims_v1(target_record_id uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION public.entities_release_active_identifier_claims_v1(target_record_id uuid) TO cartulary_runtime;

-- +goose StatementBegin
CREATE FUNCTION public.entities_active_identifier_claims_are_valid_v1()
 RETURNS boolean
 LANGUAGE sql
 STABLE SECURITY DEFINER
 SET search_path = pg_catalog, public
AS $$
    WITH expected AS (
        SELECT *
          FROM public.entities_expected_active_identifier_claims_v1(NULL::uuid)
    ), difference AS (
        (SELECT * FROM expected
         EXCEPT
         SELECT incident_id, entity_type, identifier_type,
                normalized_value, record_id
           FROM public.entity_active_identifier_claims)
        UNION ALL
        (SELECT incident_id, entity_type, identifier_type,
                normalized_value, record_id
           FROM public.entity_active_identifier_claims
         EXCEPT
         SELECT * FROM expected)
    )
    SELECT NOT EXISTS (SELECT 1 FROM difference)
    AND NOT EXISTS (
        SELECT 1 FROM public.hosts h FULL JOIN public.records r
          ON r.record_id=h.record_id AND r.record_type='host'
        LEFT JOIN public.hosts target ON target.record_id=h.merged_into_record_id
        WHERE (h.record_id IS NOT NULL OR r.record_type='host') AND (
          h.record_id IS NULL OR r.record_id IS NULL
          OR (h.incident_id,h.row_version,h.created_at,h.updated_at,h.created_by_user_id,h.updated_by_user_id)
             IS DISTINCT FROM (r.incident_id,r.row_version,r.created_at,r.updated_at,r.created_by_user_id,r.updated_by_user_id)
          OR (h.host_state='merged' AND (target.record_id IS NULL OR target.record_id=h.record_id
             OR target.incident_id IS DISTINCT FROM h.incident_id OR target.host_state='merged')))
    )
    AND NOT EXISTS (
        SELECT 1 FROM public.identities i FULL JOIN public.records r
          ON r.record_id=i.record_id AND r.record_type='identity'
        LEFT JOIN public.identities target ON target.record_id=i.merged_into_record_id
        WHERE (i.record_id IS NOT NULL OR r.record_type='identity') AND (
          i.record_id IS NULL OR r.record_id IS NULL
          OR (i.incident_id,i.row_version,i.created_at,i.updated_at,i.created_by_user_id,i.updated_by_user_id)
             IS DISTINCT FROM (r.incident_id,r.row_version,r.created_at,r.updated_at,r.created_by_user_id,r.updated_by_user_id)
          OR (i.identity_state='merged' AND (target.record_id IS NULL OR target.record_id=i.record_id
             OR target.incident_id IS DISTINCT FROM i.incident_id OR target.identity_state='merged')))
    )
    AND NOT EXISTS (
        SELECT 1 FROM public.entity_aliases a
        WHERE a.raw_text IS DISTINCT FROM normalize(public.entities_trim_unicode_space_v1(a.raw_text), NFC)
           OR a.normalized_text::text IS DISTINCT FROM a.raw_text
           OR (a.entity_type='host' AND NOT EXISTS (SELECT 1 FROM public.hosts h WHERE h.record_id=a.record_id AND h.incident_id=a.incident_id))
           OR (a.entity_type='identity' AND NOT EXISTS (SELECT 1 FROM public.identities i WHERE i.record_id=a.record_id AND i.incident_id=a.incident_id))
    )
    AND NOT EXISTS (
        SELECT 1 FROM public.entity_preserved_identifiers p
        WHERE public.entities_normalize_identifier_v1(p.identifier_type,p.raw_value) IS NULL
           OR p.normalized_value IS DISTINCT FROM public.entities_normalize_identifier_v1(p.identifier_type,p.raw_value)
           OR NOT ((p.entity_type='host' AND p.identifier_type IN ('aad_device_id','fqdn','hostname'))
              OR (p.entity_type='identity' AND p.identifier_type IN ('aad_object_id','sid','upn','email','sam_account_name')))
           OR (p.entity_type='host' AND NOT EXISTS (SELECT 1 FROM public.hosts h WHERE h.record_id=p.record_id AND h.incident_id=p.incident_id))
           OR (p.entity_type='identity' AND NOT EXISTS (SELECT 1 FROM public.identities i WHERE i.record_id=p.record_id AND i.incident_id=p.incident_id))
    )
    AND NOT EXISTS (
        SELECT 1 FROM public.entity_mentions m
        JOIN public.records source ON source.record_id=m.source_record_id
        LEFT JOIN public.records target ON target.record_id=m.resolved_record_id
        WHERE m.resolution_status='resolved' AND (target.record_id IS NULL
          OR target.incident_id IS DISTINCT FROM source.incident_id
          OR target.record_type IS DISTINCT FROM m.entity_type)
    )
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.entities_active_identifier_claims_are_valid_v1() FROM PUBLIC;

GRANT EXECUTE ON FUNCTION public.entities_active_identifier_claims_are_valid_v1() TO cartulary_recovery;

-- +goose StatementBegin
CREATE FUNCTION public.entities_rebuild_active_identifier_claims_v1()
 RETURNS bigint
 LANGUAGE plpgsql
 SECURITY DEFINER
 SET search_path = pg_catalog, public
AS $$
DECLARE
    rebuilt_count bigint;
BEGIN
    LOCK TABLE public.records IN SHARE MODE;
    LOCK TABLE public.hosts IN SHARE MODE;
    LOCK TABLE public.identities IN SHARE MODE;
    LOCK TABLE public.entity_preserved_identifiers IN SHARE MODE;
    LOCK TABLE public.entity_active_identifier_claims IN EXCLUSIVE MODE;
    DELETE FROM public.entity_active_identifier_claims;
    INSERT INTO public.entity_active_identifier_claims (
        incident_id,
        entity_type,
        identifier_type,
        normalized_value,
        record_id
    )
    SELECT incident_id, entity_type, identifier_type,
           normalized_value, record_id
      FROM public.entities_expected_active_identifier_claims_v1(NULL::uuid)
     ORDER BY incident_id, entity_type, identifier_type, normalized_value;
    GET DIAGNOSTICS rebuilt_count = ROW_COUNT;
    RETURN rebuilt_count;
END;
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.entities_rebuild_active_identifier_claims_v1() FROM PUBLIC;

GRANT EXECUTE ON FUNCTION public.entities_rebuild_active_identifier_claims_v1() TO cartulary_recovery;

-- +goose StatementBegin
CREATE FUNCTION public.entities_refresh_active_identifier_claims_v1(target_record_id uuid)
 RETURNS bigint
 LANGUAGE plpgsql
 SECURITY DEFINER
 SET search_path = pg_catalog, public
AS $$
DECLARE
    refreshed_count bigint;
BEGIN
    DELETE FROM public.entity_active_identifier_claims
     WHERE record_id = target_record_id;
    INSERT INTO public.entity_active_identifier_claims (
        incident_id,
        entity_type,
        identifier_type,
        normalized_value,
        record_id
    )
    SELECT expected.incident_id, expected.entity_type,
           expected.identifier_type, expected.normalized_value,
           expected.record_id
      FROM public.entities_expected_active_identifier_claims_v1(
          target_record_id
      ) expected
     ORDER BY expected.incident_id, expected.entity_type,
              expected.identifier_type, expected.normalized_value;
    GET DIAGNOSTICS refreshed_count = ROW_COUNT;
    RETURN refreshed_count;
END;
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.entities_refresh_active_identifier_claims_v1(target_record_id uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION public.entities_refresh_active_identifier_claims_v1(target_record_id uuid) TO cartulary_runtime;

-- +goose StatementBegin
CREATE FUNCTION public.entities_sync_active_identifier_claims_v1()
 RETURNS trigger
 LANGUAGE plpgsql
 SECURITY DEFINER
 SET search_path = pg_catalog, public
AS $$
BEGIN
	IF pg_catalog.current_setting(
		'cartulary.entities_defer_active_identifier_claims',
		true
	) = 'on' THEN
		RETURN COALESCE(NEW, OLD);
	END IF;
    IF TG_OP = 'DELETE' THEN
        PERFORM public.entities_refresh_active_identifier_claims_v1(
            OLD.record_id
        );
        RETURN OLD;
    END IF;
    IF TG_OP = 'UPDATE' AND OLD.record_id IS DISTINCT FROM NEW.record_id THEN
        PERFORM public.entities_refresh_active_identifier_claims_v1(
            OLD.record_id
        );
    END IF;
    PERFORM public.entities_refresh_active_identifier_claims_v1(
        NEW.record_id
    );
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.entities_sync_active_identifier_claims_v1() FROM PUBLIC;

CREATE INDEX entity_active_identifier_claims_record_envelope_idx ON public.entity_active_identifier_claims USING btree (incident_id, record_id);

CREATE INDEX entity_active_identifier_claims_record_lookup_idx ON public.entity_active_identifier_claims USING btree (record_id, entity_type, identifier_type, normalized_value);

CREATE INDEX entity_active_identifier_claims_typed_record_idx ON public.entity_active_identifier_claims (incident_id, record_id, entity_type);

CREATE INDEX entity_aliases_created_by_user_id_fk_idx ON public.entity_aliases USING btree (created_by_user_id);

CREATE INDEX entity_aliases_incident_id_fk_idx ON public.entity_aliases USING btree (incident_id);

CREATE INDEX entity_aliases_lookup_idx ON public.entity_aliases USING btree (incident_id, entity_type, normalized_text, record_id) WHERE (deleted_at IS NULL);

CREATE UNIQUE INDEX entity_aliases_record_unique_idx ON public.entity_aliases USING btree (record_id, entity_type, normalized_text) WHERE (deleted_at IS NULL);

CREATE INDEX entity_aliases_typed_record_idx ON public.entity_aliases (incident_id, record_id, entity_type);

CREATE INDEX entity_mentions_created_by_user_id_fk_idx ON public.entity_mentions USING btree (created_by_user_id);

CREATE INDEX entity_mentions_resolved_by_user_id_fk_idx ON public.entity_mentions USING btree (resolved_by_user_id);

CREATE INDEX entity_mentions_source_lookup_idx ON public.entity_mentions USING btree (source_record_id, source_field_key, ordinal, entity_mention_id);

CREATE INDEX entity_mentions_unresolved_lookup_idx ON public.entity_mentions USING btree (source_record_id, resolution_status, entity_type);

CREATE INDEX entity_preserved_identifiers_created_by_user_id_fk_idx ON public.entity_preserved_identifiers USING btree (created_by_user_id);

CREATE INDEX entity_preserved_identifiers_incident_id_fk_idx ON public.entity_preserved_identifiers USING btree (incident_id);

CREATE UNIQUE INDEX entity_preserved_identifiers_record_unique_idx ON public.entity_preserved_identifiers USING btree (record_id, entity_type, identifier_type, normalized_value, classification) WHERE (deleted_at IS NULL);

CREATE INDEX entity_preserved_identifiers_typed_record_idx ON public.entity_preserved_identifiers (incident_id, record_id, entity_type);

CREATE INDEX hosts_created_by_user_id_fk_idx ON public.hosts USING btree (created_by_user_id);

CREATE INDEX hosts_incident_display_name_idx ON public.hosts USING btree (incident_id, display_name, record_id);

CREATE INDEX hosts_incident_id_record_id_fk_idx ON public.hosts USING btree (incident_id, record_id);

CREATE INDEX hosts_incident_merged_into_idx ON public.hosts USING btree (incident_id, merged_into_record_id, record_id) WHERE (merged_into_record_id IS NOT NULL);

CREATE INDEX hosts_merged_into_record_id_fk_idx ON public.hosts USING btree (merged_into_record_id);

CREATE INDEX hosts_merged_into_record_id_scope_idx ON public.hosts (incident_id, merged_into_record_id);

CREATE INDEX hosts_seed_entity_mention_id_fk_idx ON public.hosts USING btree (seed_entity_mention_id);

CREATE INDEX hosts_updated_by_user_id_fk_idx ON public.hosts USING btree (updated_by_user_id);

CREATE INDEX identities_created_by_user_id_fk_idx ON public.identities USING btree (created_by_user_id);

CREATE INDEX identities_incident_display_name_idx ON public.identities USING btree (incident_id, display_name, record_id);

CREATE INDEX identities_incident_id_record_id_fk_idx ON public.identities USING btree (incident_id, record_id);

CREATE INDEX identities_incident_merged_into_idx ON public.identities USING btree (incident_id, merged_into_record_id, record_id) WHERE (merged_into_record_id IS NOT NULL);

CREATE INDEX identities_merged_into_record_id_fk_idx ON public.identities USING btree (merged_into_record_id);

CREATE INDEX identities_merged_into_record_id_scope_idx ON public.identities (incident_id, merged_into_record_id);

CREATE INDEX identities_seed_entity_mention_id_fk_idx ON public.identities USING btree (seed_entity_mention_id);

CREATE INDEX identities_updated_by_user_id_fk_idx ON public.identities USING btree (updated_by_user_id);

ALTER TABLE public.entity_active_identifier_claims ADD CONSTRAINT entity_active_identifier_claims_record_envelope_fkey FOREIGN KEY (incident_id, record_id) REFERENCES public.records(incident_id, record_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.entity_active_identifier_claims ADD CONSTRAINT entity_active_identifier_claims_typed_record_fkey FOREIGN KEY (incident_id, record_id, entity_type) REFERENCES public.records(incident_id, record_id, record_type) ON UPDATE NO ACTION ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE public.entity_aliases ADD CONSTRAINT entity_aliases_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.entity_aliases ADD CONSTRAINT entity_aliases_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.entity_aliases ADD CONSTRAINT entity_aliases_typed_record_fkey FOREIGN KEY (incident_id, record_id, entity_type) REFERENCES public.records(incident_id, record_id, record_type) ON UPDATE NO ACTION ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE public.entity_mentions ADD CONSTRAINT entity_mentions_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.entity_mentions ADD CONSTRAINT entity_mentions_resolved_by_user_id_fkey FOREIGN KEY (resolved_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.entity_mentions ADD CONSTRAINT entity_mentions_source_record_envelope_fkey FOREIGN KEY (source_record_id) REFERENCES public.records(record_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.entity_preserved_identifiers ADD CONSTRAINT entity_preserved_identifiers_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.entity_preserved_identifiers ADD CONSTRAINT entity_preserved_identifiers_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.entity_preserved_identifiers ADD CONSTRAINT entity_preserved_identifiers_typed_record_fkey FOREIGN KEY (incident_id, record_id, entity_type) REFERENCES public.records(incident_id, record_id, record_type) ON UPDATE NO ACTION ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE public.hosts ADD CONSTRAINT hosts_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.hosts ADD CONSTRAINT hosts_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.hosts ADD CONSTRAINT hosts_merged_into_record_id_fkey FOREIGN KEY (incident_id, merged_into_record_id) REFERENCES public.hosts(incident_id, record_id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.hosts ADD CONSTRAINT hosts_record_envelope_fkey FOREIGN KEY (incident_id, record_id) REFERENCES public.records(incident_id, record_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.hosts ADD CONSTRAINT hosts_seed_entity_mention_id_fkey FOREIGN KEY (seed_entity_mention_id) REFERENCES public.entity_mentions(entity_mention_id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.hosts ADD CONSTRAINT hosts_updated_by_user_id_fkey FOREIGN KEY (updated_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.identities ADD CONSTRAINT identities_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.identities ADD CONSTRAINT identities_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.identities ADD CONSTRAINT identities_merged_into_record_id_fkey FOREIGN KEY (incident_id, merged_into_record_id) REFERENCES public.identities(incident_id, record_id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.identities ADD CONSTRAINT identities_record_envelope_fkey FOREIGN KEY (incident_id, record_id) REFERENCES public.records(incident_id, record_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.identities ADD CONSTRAINT identities_seed_entity_mention_id_fkey FOREIGN KEY (seed_entity_mention_id) REFERENCES public.entity_mentions(entity_mention_id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.identities ADD CONSTRAINT identities_updated_by_user_id_fkey FOREIGN KEY (updated_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

CREATE TRIGGER entities_assert_entity_alias BEFORE INSERT OR UPDATE ON public.entity_aliases FOR EACH ROW EXECUTE FUNCTION public.entities_assert_entity_child_v1();

CREATE TRIGGER entities_assert_entity_mention BEFORE INSERT OR UPDATE ON public.entity_mentions FOR EACH ROW EXECUTE FUNCTION public.entities_assert_entity_child_v1();

CREATE TRIGGER entities_assert_entity_preserved_identifier BEFORE INSERT OR UPDATE ON public.entity_preserved_identifiers FOR EACH ROW EXECUTE FUNCTION public.entities_assert_entity_child_v1();

CREATE TRIGGER entity_preserved_identifiers_sync_active_claims AFTER INSERT OR DELETE OR UPDATE ON public.entity_preserved_identifiers FOR EACH ROW EXECUTE FUNCTION public.entities_sync_active_identifier_claims_v1();

ALTER TABLE public.entity_preserved_identifiers ENABLE ALWAYS TRIGGER entity_preserved_identifiers_sync_active_claims;

CREATE TRIGGER entities_assert_host_source BEFORE INSERT OR UPDATE ON public.hosts FOR EACH ROW EXECUTE FUNCTION public.entities_assert_entity_source_v1('host');

CREATE CONSTRAINT TRIGGER entities_reject_host_source_delete AFTER DELETE ON public.hosts DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION public.entities_reject_entity_source_delete_v1('host');

CREATE TRIGGER hosts_sync_active_identifier_claims AFTER INSERT OR DELETE OR UPDATE ON public.hosts FOR EACH ROW EXECUTE FUNCTION public.entities_sync_active_identifier_claims_v1();

ALTER TABLE public.hosts ENABLE ALWAYS TRIGGER hosts_sync_active_identifier_claims;

CREATE TRIGGER entities_assert_identity_source BEFORE INSERT OR UPDATE ON public.identities FOR EACH ROW EXECUTE FUNCTION public.entities_assert_entity_source_v1('identity');

CREATE CONSTRAINT TRIGGER entities_reject_identity_source_delete AFTER DELETE ON public.identities DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION public.entities_reject_entity_source_delete_v1('identity');

CREATE TRIGGER identities_sync_active_identifier_claims AFTER INSERT OR DELETE OR UPDATE ON public.identities FOR EACH ROW EXECUTE FUNCTION public.entities_sync_active_identifier_claims_v1();

ALTER TABLE public.identities ENABLE ALWAYS TRIGGER identities_sync_active_identifier_claims;

CREATE CONSTRAINT TRIGGER entities_assert_entity_envelope_update AFTER UPDATE OF incident_id, record_type, row_version, created_at, updated_at, created_by_user_id, updated_by_user_id ON public.records DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION public.entities_assert_entity_envelope_update_v1();

CREATE TRIGGER records_sync_entity_active_identifier_claims AFTER INSERT OR DELETE OR UPDATE OF record_id, incident_id, record_type, deleted_at ON public.records FOR EACH ROW EXECUTE FUNCTION public.entities_sync_active_identifier_claims_v1();

ALTER TABLE public.records ENABLE ALWAYS TRIGGER records_sync_entity_active_identifier_claims;

-- +goose Down
-- Disposable harness cleanup only; no production downgrade or restore guarantee.
DROP TRIGGER records_sync_entity_active_identifier_claims ON public.records;
DROP TRIGGER entities_assert_entity_envelope_update ON public.records;
DROP TRIGGER identities_sync_active_identifier_claims ON public.identities;
DROP TRIGGER entities_reject_identity_source_delete ON public.identities;
DROP TRIGGER entities_assert_identity_source ON public.identities;
DROP TRIGGER hosts_sync_active_identifier_claims ON public.hosts;
DROP TRIGGER entities_reject_host_source_delete ON public.hosts;
DROP TRIGGER entities_assert_host_source ON public.hosts;
DROP TRIGGER entity_preserved_identifiers_sync_active_claims ON public.entity_preserved_identifiers;
DROP TRIGGER entities_assert_entity_preserved_identifier ON public.entity_preserved_identifiers;
DROP TRIGGER entities_assert_entity_mention ON public.entity_mentions;
DROP TRIGGER entities_assert_entity_alias ON public.entity_aliases;
ALTER TABLE public.identities DROP CONSTRAINT identities_updated_by_user_id_fkey;
ALTER TABLE public.identities DROP CONSTRAINT identities_seed_entity_mention_id_fkey;
ALTER TABLE public.identities DROP CONSTRAINT identities_record_envelope_fkey;
ALTER TABLE public.identities DROP CONSTRAINT identities_merged_into_record_id_fkey;
ALTER TABLE public.identities DROP CONSTRAINT identities_incident_id_fkey;
ALTER TABLE public.identities DROP CONSTRAINT identities_created_by_user_id_fkey;
ALTER TABLE public.hosts DROP CONSTRAINT hosts_updated_by_user_id_fkey;
ALTER TABLE public.hosts DROP CONSTRAINT hosts_seed_entity_mention_id_fkey;
ALTER TABLE public.hosts DROP CONSTRAINT hosts_record_envelope_fkey;
ALTER TABLE public.hosts DROP CONSTRAINT hosts_merged_into_record_id_fkey;
ALTER TABLE public.hosts DROP CONSTRAINT hosts_incident_id_fkey;
ALTER TABLE public.hosts DROP CONSTRAINT hosts_created_by_user_id_fkey;
ALTER TABLE public.entity_preserved_identifiers DROP CONSTRAINT entity_preserved_identifiers_typed_record_fkey;
ALTER TABLE public.entity_preserved_identifiers DROP CONSTRAINT entity_preserved_identifiers_incident_id_fkey;
ALTER TABLE public.entity_preserved_identifiers DROP CONSTRAINT entity_preserved_identifiers_created_by_user_id_fkey;
ALTER TABLE public.entity_mentions DROP CONSTRAINT entity_mentions_source_record_envelope_fkey;
ALTER TABLE public.entity_mentions DROP CONSTRAINT entity_mentions_resolved_by_user_id_fkey;
ALTER TABLE public.entity_mentions DROP CONSTRAINT entity_mentions_created_by_user_id_fkey;
ALTER TABLE public.entity_aliases DROP CONSTRAINT entity_aliases_typed_record_fkey;
ALTER TABLE public.entity_aliases DROP CONSTRAINT entity_aliases_incident_id_fkey;
ALTER TABLE public.entity_aliases DROP CONSTRAINT entity_aliases_created_by_user_id_fkey;
ALTER TABLE public.entity_active_identifier_claims DROP CONSTRAINT entity_active_identifier_claims_typed_record_fkey;
ALTER TABLE public.entity_active_identifier_claims DROP CONSTRAINT entity_active_identifier_claims_record_envelope_fkey;
DROP FUNCTION public.entities_sync_active_identifier_claims_v1();
DROP FUNCTION public.entities_refresh_active_identifier_claims_v1(target_record_id uuid);
DROP FUNCTION public.entities_rebuild_active_identifier_claims_v1();
DROP FUNCTION public.entities_active_identifier_claims_are_valid_v1();
DROP FUNCTION public.entities_release_active_identifier_claims_v1(target_record_id uuid);
DROP FUNCTION public.entities_reject_entity_source_delete_v1();
DROP FUNCTION public.entities_expected_active_identifier_claims_v1(target_record_id uuid);
DROP FUNCTION public.entities_assert_entity_source_v1();
DROP FUNCTION public.entities_assert_entity_envelope_update_v1();
DROP FUNCTION public.entities_assert_entity_child_v1();
DROP TABLE public.identities;
DROP TABLE public.hosts;
DROP TABLE public.entity_preserved_identifiers;
DROP TABLE public.entity_mentions;
DROP TABLE public.entity_aliases;
DROP TABLE public.entity_active_identifier_claims;
DROP FUNCTION public.entities_normalize_identifier_v1(identifier_type text, raw_value text);
DROP FUNCTION public.entities_trim_unicode_space_v1(candidate text);
DROP FUNCTION public.entities_source_codepoints_admitted_v1(candidate text, reject_format boolean);
