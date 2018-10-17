syn keyword hikeInitiator goal artifact file artifacts pipeline exec each regex scandir tree
syn keyword hikeOption label name key base loud suffixIsDestination rebaseFrom rebaseTo noCache
syn keyword hikeModifier merge
syn keyword hikeFilter files directories wildcard
syn keyword hikePlaceholder source dest
syn keyword hikeAction attain require

syn match hikeInt /[+-]\?\<[0-9]\+\>/
syn match hikeDelimiter /[{}]/

syn region hikeString start=+"+ end=+"+ skip=+\\.+ contains=hikeEscape,hikeBadEscape
syn match hikeEscape /\\\%([rntbafve\\"]\|x[0-9a-fA-F]\{2\}\|u[0-9a-fA-F]\{4\}\|U[0-9a-fA-F]\{8\}\)/ contained
syn match hikeBadEscape /\\[^rntbafve\\"xuU]/ contained

hi link hikeInitiator Type
hi link hikeOption PreProc
hi link hikeModifier PreProc
hi link hikeFilter hikeModifier
hi link hikePlaceholder PreProc
hi link hikeAction Keyword

hi link hikeInt Number
hi link hikeDelimiter Keyword
hi link hikeString String
hi link hikeEscape Special
hi link hikeBadEscape Error
