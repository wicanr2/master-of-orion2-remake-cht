// Non-destructive IDA Pro 9.4 IDC probe for the adjacent raw fighter paths
// (sub_3AC20/sub_3AD57) and the Antaran star-fortress loader/capacity chain.
// Raw names, operands and
// linear addresses are preserved; this script only writes a bounded report.

#include <idc.idc>

static line(f, ea)
{
  fprintf(f, "0x%X %s\n", ea, generate_disasm_line(ea, 0));
}

static function_for(ea)
{
  return get_func_attr(ea, FUNCATTR_START);
}

static function_dump(f, ea, label, max_insns)
{
  auto start;
  auto end;
  auto cur;
  auto next;
  auto count;
  auto mnem;
  auto target;

  if (get_func_attr(ea, FUNCATTR_START) == BADADDR)
    add_func(ea, BADADDR);
  start = get_func_attr(ea, FUNCATTR_START);
  end = get_func_attr(ea, FUNCATTR_END);
  fprintf(f, "\n[function %s]\n", label);
  fprintf(f, "requested=0x%X raw_name=%s start=0x%X end=0x%X\n",
          ea, get_name(ea, GN_VISIBLE), start, end);
  if (start == BADADDR || end == BADADDR)
  {
    fprintf(f, "status=no_function\n");
    return;
  }
  cur = start;
  count = 0;
  while (cur != BADADDR && cur < end && count < max_insns)
  {
    line(f, cur);
    mnem = print_insn_mnem(cur);
    if (mnem == "call" || mnem == "callf")
    {
      target = get_operand_value(cur, 0);
      fprintf(f, "  call_site=0x%X target=0x%X target_name=%s\n",
              cur, target, get_name(target, GN_VISIBLE));
    }
    next = next_head(cur, end);
    if (next == BADADDR || next <= cur)
      break;
    cur = next;
    count++;
  }
  if (count >= max_insns)
    fprintf(f, "... instruction limit reached\n");
}

static caller_context(f, ref, fn_end, context)
{
  auto cur;
  auto n;
  auto end;
  fprintf(f, "  caller_site=0x%X caller_name=%s\n", ref, get_name(ref, GN_VISIBLE));
  cur = prev_head(ref, 0);
  n = 0;
  while (cur != BADADDR && n < context)
  {
    cur = prev_head(cur, 0);
    if (cur == BADADDR)
      break;
    n++;
  }
  end = fn_end;
  cur = (n > 0) ? cur : ref;
  n = 0;
  while (cur != BADADDR && n < context * 2 + 1 && cur < end)
  {
    line(f, cur);
    cur = next_head(cur, end);
    n++;
  }
}

static callers(f, ea, label, context)
{
  auto ref;
  auto count;
  auto start;
  auto end;
  fprintf(f, "\n[callers %s]\n", label);
  fprintf(f, "target=0x%X target_name=%s\n", ea, get_name(ea, GN_VISIBLE));
  ref = get_first_cref_to(ea);
  count = 0;
  while (ref != BADADDR && count < 80)
  {
    start = get_func_attr(ref, FUNCATTR_START);
    end = get_func_attr(ref, FUNCATTR_END);
    fprintf(f, "caller_function_start=0x%X caller_function_end=0x%X\n", start, end);
    if (start != BADADDR && end != BADADDR)
      caller_context(f, ref, end, context);
    else
      line(f, ref);
    ref = get_next_cref_to(ea, ref);
    count++;
  }
  if (ref != BADADDR)
    fprintf(f, "... caller limit reached\n");
}

static data_dump(f, ea, label, size, unit)
{
  auto off;
  auto value;
  fprintf(f, "\n[data %s]\n", label);
  fprintf(f, "ea=0x%X name=%s size=0x%X unit=%d\n", ea, get_name(ea, GN_VISIBLE), size, unit);
  fprintf(f, "bytes=");
  for (off = 0; off < size; off++)
    fprintf(f, "%02X", get_wide_byte(ea + off));
  fprintf(f, "\n");
  for (off = 0; off < size; off = off + unit)
  {
    if (unit == 1)
      value = get_wide_byte(ea + off);
    else if (unit == 2)
      value = get_wide_word(ea + off);
    else
      value = get_wide_dword(ea + off);
    fprintf(f, "  +0x%X = 0x%X\n", off, value);
  }
}

static data_refs(f, ea, label)
{
  auto ref;
  auto count;
  fprintf(f, "\n[data_xrefs %s]\n", label);
  fprintf(f, "ea=0x%X name=%s\n", ea, get_name(ea, GN_VISIBLE));
  ref = get_first_dref_to(ea);
  count = 0;
  while (ref != BADADDR && count < 120)
  {
    fprintf(f, "  dref=0x%X function_start=0x%X %s\n",
            ref, get_func_attr(ref, FUNCATTR_START), generate_disasm_line(ref, 0));
    ref = get_next_dref_to(ea, ref);
    count++;
  }
  if (ref != BADADDR)
    fprintf(f, "... data xref limit reached\n");
}

static raw_bytes(f, ea, size, label)
{
  auto off;
  fprintf(f, "\n[raw_bytes %s]\n", label);
  fprintf(f, "ea=0x%X size=0x%X bytes=", ea, size);
  for (off = 0; off < size; off++)
    fprintf(f, "%02X", get_wide_byte(ea + off));
  fprintf(f, "\n");
}

static strided_words(f, ea, count, stride, label)
{
  auto i;
  fprintf(f, "\n[strided_words %s]\n", label);
  fprintf(f, "ea=0x%X count=%d stride=0x%X address_basis=IDA_linear\n", ea, count, stride);
  for (i = 0; i < count; i++)
    fprintf(f, "  index=%d ea=0x%X word0=0x%X word2=0x%X\n",
            i, ea + i * stride, get_wide_word(ea + i * stride),
            get_wide_word(ea + i * stride + 2));
}

static weapon_record(f, id)
{
  auto off;
  off = 0x17F80D + id * 0x1C;
  fprintf(f, "  id=%d field_base=0x%X byte_plus2=0x%X word_plus4=0x%X word_plus8=0x%X word_plusA=0x%X word_plusC=0x%X\n",
          id, off, get_wide_byte(off + 2), get_wide_word(off + 4),
          get_wide_word(off + 8), get_wide_word(off + 0xA),
          get_wide_word(off + 0xC));
  fprintf(f, "  id=%d raw_record_bytes=", id);
  for (off = 0; off < 0x1C; off++)
    fprintf(f, "%02X", get_wide_byte(0x17F807 + id * 0x1C + off));
  fprintf(f, "\n");
}

static weapon_records(f)
{
  fprintf(f, "\n[weapon_records_used_by_antaran_and_fighter]\n");
  fprintf(f, "table_field_base=0x17F80D stride=0x1C address_basis=IDA_linear\n");
  weapon_record(f, 4);
  weapon_record(f, 11);
  weapon_record(f, 13);
  weapon_record(f, 24);
  weapon_record(f, 31);
  weapon_record(f, 37);
}

static dump(f, ea, label, insns)
{
  function_dump(f, ea, label, insns);
  callers(f, ea, label, 12);
}

static main()
{
  auto f;
  f = fopen("/host-tmp/moo2-fire-fortress-ida.txt", "w");
  fprintf(f, "MOO2 DEEP FIRE/FORTRESS IDA PROBE\n");
  fprintf(f, "tool=IDA Pro 9.4 idat IDC; static_only=true; runtime_not_executed\n");
  fprintf(f, "input_path=%s\n", get_input_file_path());
  fprintf(f, "input_sha256=recorded_by_container_command\n");
  fprintf(f, "address_basis=IDA linear addresses; raw names/operands preserved\n");
  fprintf(f, "min_ea=0x%X max_ea=0x%X\n", get_inf_attr(INF_MIN_EA), get_inf_attr(INF_MAX_EA));

  fprintf(f, "\n[fire_fighter_bomb_and_downstream]\n");
  dump(f, 0x3AC20, "raw_sub_3AC20_Fire_Fighter_Bomb_candidate", 900);
  dump(f, 0x3AD57, "raw_sub_3AD57_Fire_Fighter_Beam_candidate", 1200);
  dump(f, 0x3DF8D, "raw_sub_3DF8D_fighter_downstream", 500);
  dump(f, 0x3DFE0, "raw_sub_3DFE0_fighter_downstream", 500);
  dump(f, 0x36266, "raw_sub_36266_fighter_downstream", 500);
  dump(f, 0x5680D, "raw_sub_5680D_fighter_downstream", 500);
  dump(f, 0x35EAE, "raw_sub_35EAE_fighter_downstream", 500);
  dump(f, 0x1247A0, "raw_sub_1247A0_random_helper", 400);
  dump(f, 0x3CD21, "raw_sub_3CD21_fighter_speed_or_ocv", 700);
  dump(f, 0x3C892, "raw_sub_3C892_fighter_runtime_copy", 700);
  dump(f, 0x39985, "raw_sub_39985_damage_consumer", 800);
  dump(f, 0x3A0B9, "raw_sub_3A0B9_damage_consumer", 800);
  dump(f, 0x3E095, "raw_sub_3E095_missile_or_fighter_dcv", 700);
  dump(f, 0x2B7CC, "raw_sub_2B7CC_fighter_callsite", 700);
  dump(f, 0x3D839, "raw_sub_3D839_fighter_callsite", 700);
  dump(f, 0x3D884, "raw_sub_3D884_fighter_callsite", 700);
  dump(f, 0x3D2DF, "raw_sub_3D2DF_fighter_callsite", 700);
  dump(f, 0x38B5E, "raw_sub_38B5E_fighter_callsite", 700);
  raw_bytes(f, 0x3AD9C, 0xB5, "raw_sub_3AD57_flag_and_damage_slice");

  fprintf(f, "\n[antaran_fortress_and_capacity]\n");
  dump(f, 0x40148, "raw_sub_40148_combat_defender_setup", 900);
  dump(f, 0x4D18E, "raw_sub_4D18E_antaran_fortress_loader", 1800);
  dump(f, 0x6EE8E, "raw_sub_6EE8E_fortress_capacity_candidate", 700);
  dump(f, 0x6EFEB, "raw_sub_6EFEB_fortress_helper_candidate", 500);
  dump(f, 0x6A636, "raw_sub_6A636_fortress_helper_candidate", 500);
  dump(f, 0x6A406, "raw_sub_6A406_fortress_helper_candidate", 500);
  dump(f, 0x6F11C, "raw_sub_6F11C_fortress_helper_candidate", 500);
  dump(f, 0x6D18B, "raw_sub_6D18B_shared_return_tail", 180);
  dump(f, 0x78D3E, "raw_sub_78D3E_weapon_classification_helper", 500);
  dump(f, 0x6E70A, "raw_sub_6E70A_fortress_helper_candidate", 500);
  dump(f, 0x6D048, "raw_sub_6D048_tech_level_helper", 500);
  dump(f, 0x6E60E, "raw_sub_6E60E_fortress_helper_candidate", 500);
  dump(f, 0x8E94D, "raw_sub_8E94D_weapon_modifier_value_helper", 500);
  dump(f, 0x6F1CC, "raw_sub_6F1CC_weapon_modifier_helper", 500);
  dump(f, 0x127712, "raw_sub_127712_record_fill_helper", 300);
  dump(f, 0x127776, "raw_sub_127776_slot_fill_helper", 300);

  data_dump(f, 0x192864, "global_design_table_candidate", 0x139 * 13, 1);
  data_refs(f, 0x192864, "global_design_table_candidate");
  data_dump(f, 0x19917A, "antaran_size_counts_candidate", 0x20, 2);
  data_refs(f, 0x19917A, "antaran_size_counts_candidate");
  data_dump(f, 0x180140, "fortress_constant_block_candidate", 0x30, 1);
  data_refs(f, 0x180140, "fortress_constant_block_candidate");
  data_dump(f, 0x19988E, "fortress_global_word_candidate", 0x20, 2);
  data_refs(f, 0x19988E, "fortress_global_word_candidate");
  data_dump(f, 0x17F642, "ship_class_power_table_candidate", 0x90, 2);
  data_refs(f, 0x17F642, "ship_class_power_table_candidate");
  data_dump(f, 0x17FD15, "weapon_modifier_percent_table_candidate", 0x80, 2);
  strided_words(f, 0x17FD15, 15, 0x0F, "weapon_modifier_percent_rows_used_by_sub_6A406");
  data_refs(f, 0x17FD15, "weapon_modifier_percent_table_candidate");
  data_dump(f, 0x17F80D, "weapon_table_field_base_candidate", 0x1C * 16, 1);
  data_refs(f, 0x17F80D, "weapon_table_field_base_candidate");
  data_dump(f, 0x17F807, "weapon_table_candidate", 0x1C * 40, 1);
  data_refs(f, 0x17F807, "weapon_table_candidate");
  weapon_records(f);
  fprintf(f, "\n[weapon_classification_inputs]\n");
  fprintf(f, "byte_17E085_plus_13x123=0x%X byte_17E085_plus_13x47=0x%X\n",
          get_wide_byte(0x17E085 + 13 * 123), get_wide_byte(0x17E085 + 13 * 47));

  fprintf(f, "\n[raw_table_values_used_by_fortress_probe]\n");
  fprintf(f, "word_180140=0x%X word_180142=0x%X word_180144=0x%X word_180146=0x%X word_19988E=0x%X\n",
          get_wide_word(0x180140), get_wide_word(0x180142), get_wide_word(0x180144),
          get_wide_word(0x180146), get_wide_word(0x19988E));
  fprintf(f, "weapon_id4_record_base=0x%X weapon_id11_record_base=0x%X\n",
          0x17F807 + 4 * 0x1C, 0x17F807 + 11 * 0x1C);
  // The instruction uses [word_17F642 + class*0x0F] directly.  The table
  // records are 0x0F bytes apart; do not multiply the byte stride by two
  // merely because the loaded field is a word.
  fprintf(f, "class6_power_ea=0x%X value=0x%X\n",
          0x17F642 + 6 * 0x0F, get_wide_word(0x17F642 + 6 * 0x0F));

  fclose(f);
  qexit(0);
}
