// Non-destructive IDA Pro 9.4 IDC probe for the three remaining player-facing
// consumer gaps: CMBTSHP timing, SABOTAGE score inputs, and leader ETA callback.
// The labels in this report are ledger labels only; raw names, operands and
// linear addresses remain untouched.  Run this against a writable copy of the
// private .i64 database and do not save the source database.

#include <idc.idc>

static line(f, ea)
{
  fprintf(f, "0x%X %s\n", ea, generate_disasm_line(ea, 0));
}

static dump_function(f, ea, label, max_insns)
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
  fprintf(f, "requested=0x%X raw_name=%s start=0x%X end=0x%X address_basis=IDA_linear\n",
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

static caller_context(f, ref, context)
{
  auto start;
  auto end;
  auto cur;
  auto n;

  start = get_func_attr(ref, FUNCATTR_START);
  end = get_func_attr(ref, FUNCATTR_END);
  fprintf(f, "  caller_site=0x%X caller_start=0x%X caller_name=%s\n",
          ref, start, get_name(start, GN_VISIBLE));
  if (start == BADADDR || end == BADADDR)
  {
    line(f, ref);
    return;
  }
  cur = ref;
  n = 0;
  while (n < context)
  {
    cur = prev_head(cur, start);
    if (cur == BADADDR)
      break;
    n++;
  }
  n = 0;
  while (cur != BADADDR && cur < end && n < context * 2 + 1)
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

  fprintf(f, "\n[callers %s]\n", label);
  fprintf(f, "target=0x%X target_name=%s\n", ea, get_name(ea, GN_VISIBLE));
  ref = get_first_cref_to(ea);
  count = 0;
  while (ref != BADADDR && count < 120)
  {
    caller_context(f, ref, context);
    ref = get_next_cref_to(ea, ref);
    count++;
  }
  if (ref != BADADDR)
    fprintf(f, "... caller limit reached\n");
}

static data_refs(f, ea, label)
{
  auto ref;
  auto count;

  fprintf(f, "\n[data_xrefs %s]\n", label);
  fprintf(f, "ea=0x%X name=%s address_basis=IDA_linear\n", ea, get_name(ea, GN_VISIBLE));
  ref = get_first_dref_to(ea);
  count = 0;
  while (ref != BADADDR && count < 160)
  {
    fprintf(f, "  dref=0x%X function_start=0x%X %s\n",
            ref, get_func_attr(ref, FUNCATTR_START), generate_disasm_line(ref, 0));
    ref = get_next_dref_to(ea, ref);
    count++;
  }
  if (ref != BADADDR)
    fprintf(f, "... dref limit reached\n");
}

static raw_bytes(f, ea, size, label)
{
  auto off;
  fprintf(f, "\n[raw_bytes %s]\n", label);
  fprintf(f, "ea=0x%X size=0x%X address_basis=IDA_linear bytes=", ea, size);
  for (off = 0; off < size; off++)
    fprintf(f, "%02X", get_wide_byte(ea + off));
  fprintf(f, "\n");
}

static dump_target(f, ea, label, max_insns, context)
{
  dump_function(f, ea, label, max_insns);
  callers(f, ea, label, context);
}

static main()
{
  auto f;
  f = fopen("/host-tmp/moo2-consumer-closure-ida-20260811.txt", "w");
  fprintf(f, "MOO2 CONSUMER CLOSURE STATIC IDA PROBE\n");
  fprintf(f, "tool=IDA Pro 9.4 idat IDC; static_only=true; runtime_not_executed\n");
  fprintf(f, "input_path=%s\n", get_input_file_path());
  fprintf(f, "address_basis=IDA linear addresses; raw names/operands preserved\n");
  fprintf(f, "min_ea=0x%X max_ea=0x%X\n", get_inf_attr(INF_MIN_EA), get_inf_attr(INF_MAX_EA));

  fprintf(f, "\n[cmbtshp_loader_heading_and_timer_candidates]\n");
  dump_target(f, 0x30062, "raw_sub_30062_CMBTSHP_sprite_loader", 900, 16);
  dump_target(f, 0x30631, "raw_sub_30631_CMBTSHP_draw_ship", 800, 16);
  dump_target(f, 0x31F25, "raw_sub_31F25_CMBTSHP_combat_view_draw", 800, 16);
  dump_target(f, 0x33CFA, "raw_sub_33CFA_CMBTSHP_display_view", 800, 16);
  dump_target(f, 0x49A41, "raw_sub_49A41_CMBTSHP_load_combat_ship", 1000, 16);
  dump_target(f, 0x49F99, "raw_sub_49F99_CMBTSHP_load_display_sprite", 900, 16);
  dump_target(f, 0x58697, "raw_sub_58697_ship_id_picture", 400, 16);
  dump_target(f, 0x5869B, "raw_sub_5869B_ship_picture", 400, 16);
  dump_target(f, 0x586D3, "raw_sub_586D3_ship_picture_segment", 500, 16);
  dump_target(f, 0x3F5F1, "raw_sub_3F5F1_move_ship_heading", 800, 16);
  dump_target(f, 0x3F628, "raw_sub_3F628_get_facing_heading", 700, 16);

  fprintf(f, "\n[sabotage_score_and_agent_inputs]\n");
  dump_target(f, 0x10130A, "raw_sub_10130A_Steal_App", 600, 20);
  dump_target(f, 0x1014A4, "raw_sub_1014A4_N_Spies_Bonus", 900, 20);
  dump_target(f, 0x100BC5, "raw_sub_100BC5_random_context_selector", 500, 20);
  dump_target(f, 0x10119C, "raw_sub_10119C_spy_score_caller", 700, 20);
  dump_target(f, 0x101483, "raw_sub_101483_spy_score_helper", 500, 20);
  dump_target(f, 0x1026CF, "raw_sub_1026CF_spy_storage_helper", 500, 20);
  dump_target(f, 0x1026F1, "raw_sub_1026F1_their_spy_number", 300, 16);
  dump_target(f, 0x102711, "raw_sub_102711_their_spy_mission", 300, 16);
  dump_target(f, 0x102739, "raw_sub_102739_my_spy_number", 300, 16);
  dump_target(f, 0x10275F, "raw_sub_10275F_my_spy_mission", 300, 16);
  dump_target(f, 0x102776, "raw_sub_102776_my_agent_number", 300, 16);
  dump_target(f, 0x10278D, "raw_sub_10278D_my_agent_mission_setter", 300, 16);
  dump_target(f, 0x10192B, "raw_sub_10192B_spy_turn_caller", 700, 20);
  raw_bytes(f, 0x1ACE78, 0x100, "raw_table_0x1ACE78_score_random_terms");
  raw_bytes(f, 0x1ACE7A, 0x100, "raw_table_0x1ACE7A_score_random_terms_plus2");
  raw_bytes(f, 0x17EB3D, 0x13 * 49, "raw_table_0x17EB3D_sabotage_building_records");
  raw_bytes(f, 0x199CB0, 0x20, "raw_global_0x199CB0_spy_score_context");
  data_refs(f, 0x1ACE78, "raw_table_0x1ACE78_score_random_terms");
  data_refs(f, 0x1ACE7A, "raw_table_0x1ACE7A_score_random_terms_plus2");
  data_refs(f, 0x17EB3D, "raw_table_0x17EB3D_sabotage_building_records");
  data_refs(f, 0x199CB0, "raw_global_0x199CB0_spy_score_context");
  data_refs(f, 0x197F98, "raw_global_0x197F98_empire_table_pointer");
  data_refs(f, 0x192B18, "raw_global_0x192B18_building_table_pointer");

  fprintf(f, "\n[leader_eta_callback_and_downstream]\n");
  dump_target(f, 0x934CF, "raw_sub_934CF_Deassign_Officer", 500, 24);
  dump_target(f, 0x933F2, "raw_sub_933F2_Check_Officer_Fields", 500, 20);
  dump_target(f, 0x93528, "raw_sub_93528_UI_state_reset", 300, 16);
  dump_target(f, 0xE2AB1, "raw_sub_E2AB1_colony_calculation_candidate", 1600, 24);
  dump_target(f, 0xE1D59, "raw_sub_E1D59_callback_downstream", 500, 16);
  dump_target(f, 0xDF8F0, "raw_sub_DF8F0_callback_downstream", 500, 16);
  dump_target(f, 0xE2710, "raw_sub_E2710_callback_empire_recalc", 700, 20);
  dump_target(f, 0x98F42, "raw_sub_98F42_Get_Ship_Leader_ETA", 500, 16);
  dump_target(f, 0x943A0, "raw_sub_943A0_Move_From_Limbo_To_Pool", 700, 20);
  dump_target(f, 0x9453C, "raw_sub_9453C_Move_Officer_To_Pool", 700, 20);
  dump_target(f, 0x97287, "raw_sub_97287_Set_Officer_To_Player", 700, 20);
  dump_target(f, 0xD7662, "raw_sub_D7662_Do_AI_Leaders", 900, 20);
  dump_target(f, 0xDCDAC, "raw_sub_DCDAC_Handle_Leader_At_Lost_Colony", 500, 20);
  dump_target(f, 0xE1FC7, "raw_sub_E1FC7_Leader_At_Anomaly", 500, 20);
  data_refs(f, 0x1930DC, "raw_global_0x1930DC_leader_records");
  data_refs(f, 0x19306C, "raw_global_0x19306C_colony_records");
  data_refs(f, 0x197F9C, "raw_global_0x197F9C_ship_records");

  fclose(f);
  qexit(0);
}
