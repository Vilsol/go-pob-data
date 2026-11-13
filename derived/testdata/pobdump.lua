-- usage: luajit pobdump.lua <PoB src/Data file> [uniqueModStatOrder]
-- Writes the table the file returns to stdout as JSON. With uniqueModStatOrder, the file is
-- ModItemExclusive.lua and the output is the uniqueModStatOrder Item.lua folds it into.
local path, mode = arg[1], arg[2]
local data = dofile(path)

-- src/Classes/Item.lua:64-68
local function normaliseModLine(line)
	return line:gsub("%d+%.?%d*", "#")
		:gsub("%(%-?#%-#%)", "#"):lower()
		:gsub("\n", " ")
end

if mode == "uniqueModStatOrder" then
	-- src/Classes/Item.lua:1361-1371
	local uniqueModStatOrder = { exact = { }, normalised = { } }
	for _, mod in pairs(data) do
		for index, line in ipairs(mod) do
			local exactLine = line:lower():gsub("\n", " ")
			local statLine = normaliseModLine(line)
			uniqueModStatOrder.exact[exactLine] = math.min(uniqueModStatOrder.exact[exactLine] or math.huge, mod.statOrder[index])
			uniqueModStatOrder.normalised[statLine] = math.min(uniqueModStatOrder.normalised[statLine] or math.huge, mod.statOrder[index])
		end
	end
	data = uniqueModStatOrder
end

local function quote(s)
	return '"' .. s:gsub('[%c"\\]', function(c)
		return string.format("\\u%04x", c:byte())
	end) .. '"'
end

-- isArray is whether t's keys are exactly 1..#t, however its constructor spelled them.
local function isArray(t)
	local n = 0
	for k in pairs(t) do
		if type(k) ~= "number" or k < 1 or k ~= math.floor(k) then
			return false
		end
		n = n + 1
	end
	return n == #t
end

local out = { }
local function encode(v)
	local t = type(v)
	if t == "string" then
		out[#out + 1] = quote(v)
	elseif t == "number" then
		out[#out + 1] = v == math.floor(v) and string.format("%d", v) or string.format("%.17g", v)
	elseif t == "boolean" then
		out[#out + 1] = tostring(v)
	elseif next(v) == nil then
		-- An empty table is either; JSON null decodes to an empty Go slice or map alike.
		out[#out + 1] = "null"
	elseif isArray(v) then
		out[#out + 1] = "["
		for i, e in ipairs(v) do
			if i > 1 then out[#out + 1] = "," end
			encode(e)
		end
		out[#out + 1] = "]"
	else
		out[#out + 1] = "{"
		local first = true
		for k, e in pairs(v) do
			if not first then out[#out + 1] = "," end
			first = false
			out[#out + 1] = quote(tostring(k))
			out[#out + 1] = ":"
			encode(e)
		end
		out[#out + 1] = "}"
	end
end
encode(data)
io.write(table.concat(out))
