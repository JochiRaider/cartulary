// Documentation-only rule. Product contracts and tests never consume Markdown.
module.exports = {
  names: ['RP001', 'reference-pack-requirement-declarations'],
  description: 'Reference Pack normative requirement declarations are unique',
  tags: ['requirements'],
  function: function requirementDeclarations(params, onError) {
    if (!/(?:^|[/\\])reference-pack-subsystem-nlspec\.md$/.test(params.name)) return;
    const declarations = new Map();
    for (const token of params.tokens) {
      if (token.type !== 'inline' || token.level !== 1) continue;
      const match = /^\*\*(RP-REQ-[0-9]+)\*\*(?:\s|$)/.exec(token.content);
      if (!match) continue;
      const previous = declarations.get(match[1]);
      if (previous !== undefined) {
        onError({ lineNumber: token.lineNumber, detail: `${match[1]} was already declared on line ${previous}` });
      } else declarations.set(match[1], token.lineNumber);
    }
  }
};
