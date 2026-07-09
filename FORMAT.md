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
- 17: Indoor
- 54: OpenWater

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

## Important

SSI format is not officially documented publicly. Treat this as a practical,
community-driven compatibility specification.
