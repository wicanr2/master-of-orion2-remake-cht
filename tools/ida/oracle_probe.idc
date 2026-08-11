// Non-destructive IDA Pro 9.4 IDC oracle probe.
// Keep original names, linear addresses, and operands; only write a bounded report.

#include <idc.idc>

static report_xrefs(f, ea)
{
  auto ref;
  auto count;
  count = 0;
  ref = get_first_cref_to(ea);
  while (ref != BADADDR && count < 20)
  {
    fprintf(f, "    xref=0x%X %s\n", ref, generate_disasm_line(ref, 0));
    ref = get_next_cref_to(ea, ref);
    count++;
  }
  if (ref != BADADDR)
    fprintf(f, "    ... xref limit reached\n");
}

static report_drefs(f, ea)
{
  auto ref;
  auto count;
  count = 0;
  ref = get_first_dref_to(ea);
  while (ref != BADADDR && count < 40)
  {
    fprintf(f, "    dref=0x%X %s\n", ref, generate_disasm_line(ref, 0));
    ref = get_next_dref_to(ea, ref);
    count++;
  }
  if (ref != BADADDR)
    fprintf(f, "    ... dref limit reached\n");
}

static dump_function(f, ea, max_insns)
{
  auto start;
  auto end;
  auto cur;
  auto next;
  auto count;
  start = get_func_attr(ea, FUNCATTR_START);
  end = get_func_attr(ea, FUNCATTR_END);
  if (start == BADADDR || end == BADADDR)
  {
    fprintf(f, "  no function boundary at 0x%X\n", ea);
    return;
  }
  fprintf(f, "  function_start=0x%X function_end=0x%X original_name=%s\n",
          start, end, get_func_name(start));
  cur = start;
  count = 0;
  while (cur != BADADDR && cur < end && count < max_insns)
  {
    fprintf(f, "    0x%X %s\n", cur, generate_disasm_line(cur, 0));
    next = next_head(cur, end);
    if (next == BADADDR || next <= cur)
      break;
    cur = next;
    count++;
  }
  if (count >= max_insns)
    fprintf(f, "    ... instruction limit reached\n");
}

static dump_from_address(f, ea, max_insns)
{
  auto end;
  auto cur;
  auto next;
  auto count;
  end = get_func_attr(ea, FUNCATTR_END);
  if (end == BADADDR)
  {
    fprintf(f, "  no function end for window at 0x%X\n", ea);
    return;
  }
  fprintf(f, "  window_start=0x%X function_end=0x%X\n", ea, end);
  cur = ea;
  count = 0;
  while (cur != BADADDR && cur < end && count < max_insns)
  {
    fprintf(f, "    0x%X %s\n", cur, generate_disasm_line(cur, 0));
    next = next_head(cur, end);
    if (next == BADADDR || next <= cur)
      break;
    cur = next;
    count++;
  }
  if (count >= max_insns)
    fprintf(f, "    ... instruction limit reached\n");
}

static dump_symbol(f, requested)
{
  auto ea;
  ea = get_name_ea_simple(requested);
  if (ea == BADADDR)
  {
    fprintf(f, "symbol requested=%s status=not_found\n", requested);
    return;
  }
  fprintf(f, "symbol requested=%s ea=0x%X original_name=%s\n",
          requested, ea, get_name(ea, GN_VISIBLE));
  report_xrefs(f, ea);
  dump_function(f, ea, 80);
}

static dump_address(f, label, ea)
{
  // The private database and the LE loader may leave some procedure
  // boundaries unresolved. Asking IDA to define the function at the ledger
  // address is still non-destructive to the private input: this run operates
  // on a copy in /tmp and never saves the source .i64.
  if (get_func_attr(ea, FUNCATTR_START) == BADADDR)
    add_func(ea, BADADDR);
  fprintf(f, "address label=%s ea=0x%X disasm=%s\n", label, ea, generate_disasm_line(ea, 0));
  report_xrefs(f, ea);
  dump_function(f, ea, 56);
}

static dump_deep_address(f, label, ea, max_insns)
{
  // Keep the ledger label separate from IDA's raw location.  This is used for
  // high-impact slices where the private LE objects have historically been
  // represented with more than one address basis.
  if (get_func_attr(ea, FUNCATTR_START) == BADADDR)
    add_func(ea, BADADDR);
  fprintf(f, "deep_address label=%s ea=0x%X disasm=%s\n", label, ea, generate_disasm_line(ea, 0));
  report_xrefs(f, ea);
  dump_function(f, ea, max_insns);
}

static dump_data_address(f, label, ea)
{
  fprintf(f, "data_address label=%s ea=0x%X original_name=%s disasm=%s\n",
          label, ea, get_name(ea, GN_VISIBLE), generate_disasm_line(ea, 0));
  report_drefs(f, ea);
}

static selected_name(name)
{
  return strstr(name, "VESA") != -1 || strstr(name, "Vesa") != -1 ||
         strstr(name, "DIPLOM") != -1 || strstr(name, "Diplom") != -1 ||
         strstr(name, "FIGHTER") != -1 || strstr(name, "Fighter") != -1 ||
         strstr(name, "BOMB") != -1 || strstr(name, "Bomb") != -1 ||
         strstr(name, "LEADER") != -1 || strstr(name, "Leader") != -1 ||
         strstr(name, "GAME") != -1 || strstr(name, "Game") != -1 ||
         strstr(name, "GAM") != -1 || strstr(name, "Gam") != -1 ||
         strstr(name, "CMBTSHP") != -1 || strstr(name, "ARM") != -1 ||
         strstr(name, "FST") != -1;
}

static main()
{
  auto f;
  auto ea;
  auto name;
  auto count;
  f = fopen("/host-tmp/moo2-oracle-ida-20260811.txt", "w");
  fprintf(f, "MOO2 ORACLE STATIC PROBE\n");
  fprintf(f, "tool=IDA Pro 9.4 idat IDC; evidence=static only; runtime_not_executed\n");
  // IDA 9.4 IDC exposes no portable kernel-version getter; the container
  // command line records the exact image/tool version in the companion note.
  fprintf(f, "ida_version=IDA Pro 9.4 (container image ida-pro-9.4-ver2)\n");
  fprintf(f, "input_path=%s\n", get_input_file_path());
  fprintf(f, "address_basis=IDA linear addresses; DOS LE object #1 code base 0x10000\n");
  fprintf(f, "min_ea=0x%X max_ea=0x%X\n\n", get_inf_attr(INF_MIN_EA), get_inf_attr(INF_MAX_EA));

  fprintf(f, "[exact original symbol candidates]\n");
  dump_symbol(f, "VESA_Init_");
  dump_symbol(f, "VESA.COM");
  dump_symbol(f, "Fire_Fighter_Bomb");
  dump_symbol(f, "Fire_Fighter_Bomb_");
  dump_symbol(f, "Start_Diplomacy_Music_");
  dump_symbol(f, "Play_Diplomacy_Music_");
  dump_symbol(f, "Load_Game_");
  dump_symbol(f, "Load_Leaders_");
  dump_symbol(f, "Save_Game_");
  dump_symbol(f, "Read_Game_");
  dump_symbol(f, "Write_Game_");

  fprintf(f, "\n[matching original function names]\n");
  ea = get_next_func(0);
  count = 0;
  while (ea != BADADDR && count < 120)
  {
    name = get_func_name(ea);
    if (selected_name(name))
    {
      fprintf(f, "function ea=0x%X original_name=%s\n", ea, name);
      report_xrefs(f, ea);
      count++;
    }
    ea = get_next_func(ea);
  }
  if (count >= 120)
    fprintf(f, "... matching function limit reached\n");

  fprintf(f, "\n[known IDA linear addresses from the evidence ledger]\n");
  dump_address(f, "Play_Streaming_Music_call_or_body", 0x24677);
  dump_address(f, "Play_Background_Music_call_or_body", 0x2484F);
  dump_address(f, "Play_Combat_Music_call_or_body", 0x2496C);
  // The diplomacy music assignments were recorded in the ledger as object #1
  // offsets.  Convert them explicitly to IDA linear addresses here; do not
  // hide the object-base conversion in a renamed symbol.
  dump_address(f, "Diplomacy_good_music_assignment_obj1+0x908F", 0x1908F);
  dump_from_address(f, 0x1908F, 80);
  dump_address(f, "Diplomacy_bad_music_assignment_obj1+0x90AD", 0x190AD);
  dump_from_address(f, 0x190AD, 80);
  dump_address(f, "Diplomacy_Load_New_Ambassador", 0x1B92E);
  dump_from_address(f, 0x1BD00, 80);
  dump_address(f, "Diplomacy_random_helper", 0x1247A0);
  dump_address(f, "Start_Diplomacy_Music_body_obj1+0xD0F0", 0x1D0F0);
  dump_address(f, "Diplomacy_Test_obj1+0x533F4", 0x1533F4);
  dump_address(f, "Get_Gift_Response_obj1+0x539D9", 0x1539D9);
  dump_address(f, "Fire_Fighter_Bomb_obj1+0x3AD57", 0x13AD57);
  dump_address(f, "Fighter_Ocv_obj1+0x3DFE0", 0x13DFE0);
  dump_address(f, "Fighter_speed_downstream_obj1+0x3CD21", 0x13CD21);
  dump_address(f, "Fighter_runtime_copy_obj1+0x3C892", 0x13C892);
  dump_address(f, "Fighter_callsite_obj1+0x2B7CC", 0x12B7CC);

  fprintf(f, "\n[deep enemy blueprint and combat loader addresses]\n");
  // The fixed external ledger records these as object offsets.  The explicit
  // +0x10000 labels preserve that conversion instead of silently renaming a
  // database location.
  dump_deep_address(f, "Load_Combat_Antaran_Ship_fixed+0x10000", 0x14D10E, 220);
  dump_deep_address(f, "Load_Antaran_Defense_Fleet_fixed+0x10000", 0x14D141, 260);
  dump_deep_address(f, "Load_Antaran_Star_Fortress_fixed+0x10000", 0x14D18E, 180);
  dump_deep_address(f, "Load_Antaran_Ship_Design_fixed+0x10000", 0x15514C, 220);
  dump_deep_address(f, "Load_Small_Antaran_Combat_Ship_fixed+0x10000", 0x155161, 220);
  dump_deep_address(f, "Load_Small_Antaran_Design_fixed+0x10000", 0x155364, 260);
  dump_deep_address(f, "Load_Medium_Antaran_Combat_Ship_fixed+0x10000", 0x15542C, 220);
  dump_deep_address(f, "Load_Medium_Antaran_Design_fixed+0x10000", 0x15565D, 260);
  dump_deep_address(f, "Load_Large_Antaran_Combat_Ship_fixed+0x10000", 0x155738, 220);
  dump_deep_address(f, "Load_Large_Antaran_Design_fixed+0x10000", 0x1559DC, 260);
  dump_deep_address(f, "Load_Huge_Antaran_Combat_Ship_fixed+0x10000", 0x155B12, 220);
  dump_deep_address(f, "Load_Huge_Antaran_Design_fixed+0x10000", 0x155E16, 280);
  dump_deep_address(f, "Load_Titan_Antaran_Combat_Ship_fixed+0x10000", 0x155F67, 220);
  dump_deep_address(f, "Load_Titan_Antaran_Design_fixed+0x10000", 0x1562D6, 300);
  dump_deep_address(f, "Load_Combat_Ship_fixed+0x10000", 0x14954A, 220);
  dump_deep_address(f, "Fighter_Ocv_fixed+0x10000", 0x13DF8D, 180);
  dump_deep_address(f, "Fire_Fighter_Bomb_fixed+0x10000", 0x13AC20, 220);
  dump_deep_address(f, "Fire_Fighter_Beam_fixed+0x10000", 0x13AD57, 220);

  fprintf(f, "\n[deep diplomacy threshold addresses]\n");
  dump_deep_address(f, "Diplomacy_Test_fixed+0x10000", 0x153146, 360);
  dump_deep_address(f, "Get_Gift_Response_fixed+0x10000", 0x153723, 420);
  dump_deep_address(f, "Get_Player_Diplomacy_Personality_fixed+0x10000", 0x153E96, 160);
  dump_deep_address(f, "NPC_Proposal_Rejection_Accept_fixed", 0x276E6, 300);
  dump_deep_address(f, "NPC_Proposal_Rejection_Accept_fixed+0x10000", 0x376E6, 300);

  fprintf(f, "\n[deep GAM load/import addresses]\n");
  dump_deep_address(f, "Load_Game_fixed+0x10000", 0x110E2F, 360);
  dump_deep_address(f, "Save_Game_fixed+0x10000", 0x11160B, 360);
  dump_deep_address(f, "Init_Officers_fixed+0x10000", 0x11307E, 220);
  dump_deep_address(f, "Init_Leaders_fixed+0x10000", 0x11307F, 260);
  dump_deep_address(f, "leaders_global_fixed+0x10000", 0x1A30DC, 80);
  dump_deep_address(f, "Load_Game_fixed", 0x10E2F, 160);
  dump_deep_address(f, "Save_Game_fixed", 0x1160B, 160);
  dump_deep_address(f, "Init_Leaders_fixed", 0x1307F, 160);

  fprintf(f, "\n[deep fixed-ledger low addresses]\n");
  // These are the addresses as represented by symbols_fixed.tsv and present
  // directly in the IDA database.  They are intentionally separate from the
  // legacy object1+0x10000 probes above.
  dump_deep_address(f, "Load_Combat_Antaran_Ship_fixed", 0x4D10E, 260);
  dump_deep_address(f, "Load_Antaran_Defense_Fleet_fixed", 0x4D141, 320);
  dump_deep_address(f, "Load_Antaran_Star_Fortress_fixed", 0x4D18E, 180);
  dump_deep_address(f, "Load_Antaran_Ship_Design_fixed", 0x5514C, 260);
  dump_deep_address(f, "Load_Small_Antaran_Combat_Ship_fixed", 0x55161, 280);
  dump_deep_address(f, "Load_Small_Antaran_Design_fixed", 0x55364, 300);
  dump_deep_address(f, "Load_Medium_Antaran_Combat_Ship_fixed", 0x5542C, 280);
  dump_deep_address(f, "Load_Medium_Antaran_Design_fixed", 0x5565D, 300);
  dump_deep_address(f, "Load_Large_Antaran_Combat_Ship_fixed", 0x55738, 280);
  dump_deep_address(f, "Load_Large_Antaran_Design_fixed", 0x559DC, 340);
  dump_deep_address(f, "Load_Huge_Antaran_Combat_Ship_fixed", 0x55B12, 280);
  dump_deep_address(f, "Load_Huge_Antaran_Design_fixed", 0x55E16, 340);
  dump_deep_address(f, "Load_Titan_Antaran_Combat_Ship_fixed", 0x55F67, 280);
  dump_deep_address(f, "Load_Titan_Antaran_Design_fixed", 0x562D6, 360);
  dump_deep_address(f, "Load_Combat_Ship_fixed", 0x4954A, 300);
  dump_deep_address(f, "Fighter_Ocv_fixed", 0x3DF8D, 220);
  dump_deep_address(f, "Fire_Fighter_Bomb_fixed", 0x3AC20, 260);
  dump_deep_address(f, "Fire_Fighter_Beam_fixed", 0x3AD57, 260);

  fprintf(f, "\n[deep fixed-ledger diplomacy low addresses]\n");
  dump_deep_address(f, "Diplomacy_Test_fixed", 0x53146, 440);
  dump_deep_address(f, "Get_Gift_Response_fixed", 0x53723, 520);
  dump_deep_address(f, "Get_Player_Diplomacy_Personality_fixed", 0x53E96, 180);
  dump_deep_address(f, "NPC_Proposal_Rejection_Accept_fixed", 0x276E6, 360);

  fprintf(f, "\n[deep current-input diplomacy raw addresses]\n");
  // The two external symbol ledgers disagree around this range.  These
  // probes use the raw IDA locations from the current Orion2.exe.i64 and
  // deliberately retain sub_ names rather than promoting a ledger alias.
  dump_deep_address(f, "raw_Check_Treaty_Proposal_current_0x53146", 0x53146, 440);
  dump_deep_address(f, "raw_Diplomacy_Test_current_0x533F4", 0x533F4, 520);
  dump_deep_address(f, "raw_Get_Gift_Response_current_0x539D9", 0x539D9, 620);

  fprintf(f, "\n[deep current-input fighter downstream raw addresses]\n");
  // Keep the current database's raw sub_* locations as the primary evidence.
  // The external symbol ledgers use inconsistent semantic labels in this
  // range, so each probe is intentionally named only by its raw address.
  dump_deep_address(f, "raw_Fighter_downstream_current_0x3DF8D", 0x3DF8D, 220);
  dump_deep_address(f, "raw_Fighter_downstream_current_0x3DFE0", 0x3DFE0, 220);
  dump_deep_address(f, "raw_Fighter_downstream_current_0x3AC20", 0x3AC20, 360);
  dump_deep_address(f, "raw_Fighter_downstream_current_0x3AD57", 0x3AD57, 620);
  dump_deep_address(f, "raw_Fighter_damage_apply_current_0x39985", 0x39985, 360);
  dump_deep_address(f, "raw_Fighter_damage_apply_current_0x3A0B9", 0x3A0B9, 360);
  dump_deep_address(f, "raw_Fighter_target_alive_current_0x3D299", 0x3D299, 220);
  dump_deep_address(f, "raw_Fighter_downstream_current_0x3CD21", 0x3CD21, 360);
  dump_deep_address(f, "raw_Fighter_downstream_current_0x3C892", 0x3C892, 360);
  dump_deep_address(f, "raw_Fighter_downstream_current_0x3E095", 0x3E095, 360);
  dump_deep_address(f, "raw_Fighter_callsite_current_0x3D839", 0x3D839, 260);
  dump_deep_address(f, "raw_Fighter_callsite_current_0x3D884", 0x3D884, 260);
  dump_deep_address(f, "raw_Fighter_callsite_current_0x3D2DF", 0x3D2DF, 260);
  dump_deep_address(f, "raw_Fighter_callsite_current_0x38B5E", 0x38B5E, 360);

  fprintf(f, "\n[deep current-input Antaran fortress raw address]\n");
  dump_deep_address(f, "raw_Antaran_fortress_loader_current_0x4D18E", 0x4D18E, 900);
  fprintf(f, "fortress_loader_constants address_basis=IDA linear addresses");
  fprintf(f, " word_180140=0x%X word_180142=0x%X word_180144=0x%X word_180146=0x%X word_19988E=0x%X\n",
          get_wide_word(0x180140), get_wide_byte(0x180142),
          get_wide_word(0x180144), get_wide_word(0x180146),
          get_wide_word(0x19988E));
  fprintf(f, "fortress_table_values id4_category=0x%X id4_flags=0x%X id4_min=0x%X id4_max=0x%X id4_size=0x%X",
          get_wide_byte(0x17F80F + 4 * 0x1C), get_wide_word(0x17F815 + 4 * 0x1C),
          get_wide_word(0x17F817 + 4 * 0x1C), get_wide_word(0x17F819 + 4 * 0x1C),
          get_wide_word(0x17F811 + 4 * 0x1C));
  fprintf(f, " id11_category=0x%X id11_flags=0x%X id11_min=0x%X id11_max=0x%X id11_size=0x%X",
          get_wide_byte(0x17F80F + 11 * 0x1C), get_wide_word(0x17F815 + 11 * 0x1C),
          get_wide_word(0x17F817 + 11 * 0x1C), get_wide_word(0x17F819 + 11 * 0x1C),
          get_wide_word(0x17F811 + 11 * 0x1C));
  fprintf(f, " class6_power=0x%X\n", get_wide_word(0x17F642 + 6 * 0x0F * 2));
  dump_deep_address(f, "raw_Antaran_fortress_space_cost_current_0x6EE8E", 0x6EE8E, 220);
  dump_deep_address(f, "raw_Antaran_fortress_cost_helper_current_0x6EFEB", 0x6EFEB, 180);
  dump_deep_address(f, "raw_Antaran_fortress_cost_helper_current_0x6A636", 0x6A636, 180);
  dump_deep_address(f, "raw_Antaran_fortress_cost_helper_current_0x6A406", 0x6A406, 180);
  dump_deep_address(f, "raw_Antaran_fortress_cost_helper_current_0x6F11C", 0x6F11C, 180);
  dump_deep_address(f, "raw_Antaran_fortress_cost_helper_current_0x6E70A", 0x6E70A, 180);
  dump_deep_address(f, "raw_Antaran_fortress_cost_helper_current_0x6E60E", 0x6E60E, 180);

  fprintf(f, "\n[deep fixed-ledger GAM low addresses]\n");
  dump_deep_address(f, "Load_Game_fixed", 0x10E2F, 620);
  dump_deep_address(f, "Save_Game_fixed", 0x1160B, 620);
  dump_deep_address(f, "Init_Officers_fixed", 0x1307E, 260);
  dump_deep_address(f, "Init_Leaders_fixed", 0x1307F, 300);
  dump_deep_address(f, "leaders_global_fixed", 0x1930DC, 80);

  fprintf(f, "\n[deep current-input GAM raw addresses]\n");
  // Capture the full current-input loader/save windows so the fread/fwrite
  // shape of each global block can be compared without relying on an older
  // decompilation's semantic names.
  dump_deep_address(f, "raw_Load_Game_current_0x10E2F", 0x10E2F, 1200);
  dump_deep_address(f, "raw_Save_Game_current_0x1160B", 0x1160B, 1000);
  dump_deep_address(f, "raw_Init_Leaders_current_0x1307F", 0x1307F, 340);
  dump_data_address(f, "raw_global_block_current_0x1930DC", 0x1930DC);

  fprintf(f, "\n[bounded selected function disassembly]\n");
  ea = get_next_func(0);
  count = 0;
  while (ea != BADADDR && count < 40)
  {
    name = get_func_name(ea);
    if (selected_name(name))
    {
      fprintf(f, "selected ea=0x%X original_name=%s\n", ea, name);
      dump_function(f, ea, 48);
      count++;
    }
    ea = get_next_func(ea);
  }
  fclose(f);
  qexit(0);
}
