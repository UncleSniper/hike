syn keyword hikeInitiator goal artifact file artifacts pipeline exec each regex scandir tree
syn keyword hikeInitiator delete split set setdef include copy zip piece unzip valve directory
syn keyword hikeOption label name key base loud suffixIsDestination rebaseFrom rebaseTo noCache
syn keyword hikeOption toDirectory from to rename
syn keyword hikeModifier merge ifExists
syn keyword hikeFilter files directories wildcard any all not
syn keyword hikePlaceholder source dest aux
syn keyword hikeAction attain require
syn keyword hikeSetting projectKey

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
hi link hikeSetting hikeModifier

hi link hikeInt Number
hi link hikeDelimiter Keyword
hi link hikeString String
hi link hikeEscape Special
hi link hikeBadEscape Error
