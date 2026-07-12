# SSI QR Format (Reverse-Engineered)

## Encoding

- Prefix: `dive;noid;`
- Field separator: `;`
- Key-value separator: `:`
- Date-time format: `YYYYMMDDHHmm`

## Core Required Fields

- `dive_type` (int)
- `divetime` (float, minutes)
- `datetime` (string)
- `depth_m` (float, meters)

## Common Optional Fields

- `site` (int)
- `var_weather_id` (int)
- `var_entry_id` (int)
- `var_water_body_id` (int)
- `var_watertype_id` (int)
- `var_current_id` (int)
- `var_surface_id` (int)
- `var_divetype_id` (int)
- `user_master_id` (int)
- `user_firstname` (string)
- `user_lastname` (string)
- `user_leader_id` (int)
- `airtemp_c` (float)
- `watertemp_c` (float)
- `vis_m` (float)

## Known Mappings (Public Reverse Engineering)

### dive_type

- 0: Scuba
- 2: ExtendedRange
- 4: RebreatherSelfContained
- 6: Freediving
- 8: RebreatherClosedCircuit

### var_weather_id

- 1: Cloudless
- 2: Cloudy
- 3: Rainy
- 121: Snow

### var_entry_id

- 21: ShoreOrBeach
- 22: Boat
- 35: Other

### var_water_body_id (incomplete)

- 13: Ocean
- 14: River
- 15: Quarry
- 16: Lake
- 17: Indoor — confirmed: observed in a real SSI-app-generated QR payload for
  a dive at an indoor facility ("Centrum Indoor")
- 54: OpenWater

The application resolves this field from Subsurface metadata in the following
order: a per-dive choice set via the GUI's dive list (right-click a row to
pick a category, optionally applied to every dive sharing the same site
name for the current import), an unambiguous local text match in the
site/dive metadata, then the configured fallback (CLI/library only). When
none applies, the field is omitted. It is not inferred from GPS or any
online lookup.

The local text match recognizes a few unique dedicated dive-pool brand names
as Indoor even without a generic keyword like "indoor" or "pool", e.g.
"Deepspot" (Poland).

### var_watertype_id

- 4: Fresh
- 5: Salt

### var_current_id

- 6: NoCurrent
- 7: LightCurrent
- 8: StrongCurrent
- 9: RippingCurrent

### var_surface_id

- 10: Calm
- 11: Moving
- 12: Stormy

### var_divetype_id

- 23: Education
- 24: FunDive
- 138: Scientific
- 139: Work

## Confidence Levels

- Confirmed: fields and IDs observed in public reverse-engineering sources.
- Inferred: mappings derived from behavior and naming conventions.
- Unknown: undocumented fields or IDs not publicly reproducible.

## Example Payload

```text
dive;noid;dive_type:0;divetime:48.5;datetime:202509201623;depth_m:26.4;var_weather_id:2;var_entry_id:21;var_water_body_id:15;var_watertype_id:5;var_current_id:6;var_surface_id:10;var_divetype_id:24
```

## Real-World Confirmed Example (Indoor Dive)

The following is a genuine SSI-app-generated QR payload for a dive at an
indoor facility, with personal fields (`site`, `user_master_id`,
`user_firstname`, `user_lastname`) redacted/replaced since they identify a
real person and a real SSI-internal site ID:

```text
dive;noid;dive_type:0;divetime:53.0;datetime:202412091800;depth_m:12.3;site:REDACTED;var_water_body_id:17;var_watertype_id:4;var_divetype_id:24;var_divetype_id:24;user_master_id:REDACTED;user_firstname:REDACTED;user_lastname:REDACTED;user_leader_id:
```

Notes from this real example (informational; not replicated by this
application's own serializer, since they look like quirks of the official
app's generator rather than required compatibility behavior):

- `var_divetype_id:24` appears twice.
- `user_leader_id:` is emitted with an empty value (trailing key with no
  value) rather than being omitted.
- `site` is a plain integer referencing SSI's own internal dive-site
  database, unrelated to any Subsurface identifier.

## Important

SSI format is not officially documented publicly. Treat this as a practical,
community-driven compatibility specification.
