---@meta

--[[
TODO:
https://github.com/LuaLS/lua-language-server/issues/1861
This class definition does not represent what's actually
going on under-the-hood.  gobindlua returns a UserData type
that is a thin wrapper around slices/maps, and only converts
to tables when it needs to.  This to_table method is actually
supposed to accept the gbl_array/gbl_map and return a table,
but due to the lack of generic classes, we're just representing
these gbl_array/gbl_map types as tables.
]]

---@class gbl_map
gbl_map = {}

---@generic K
---@generic V
---@param t table<K,V>
---@return table<K,V>
function gbl_map.to_table(t) end
