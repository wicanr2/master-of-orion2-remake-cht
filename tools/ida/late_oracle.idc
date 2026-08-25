// Non-destructive IDA Pro 9.4 IDC probe for late remake parity gaps.
// Raw names, operands and linear addresses are preserved.  Labels below are
// evidence-ledger labels, not replacements for original symbols.

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
    if (print_insn_mnem(cur) == "call")
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

static callers(f, ea, label)
{
  auto ref;
  auto count;
  fprintf(f, "\n[callers %s]\n", label);
  fprintf(f, "target=0x%X target_name=%s\n", ea, get_name(ea, GN_VISIBLE));
  ref = get_first_cref_to(ea);
  count = 0;
  while (ref != BADADDR && count < 100)
  {
    fprintf(f, "  caller_site=0x%X caller_start=0x%X caller_name=%s %s\n",
            ref, get_func_attr(ref, FUNCATTR_START),
            get_name(get_func_attr(ref, FUNCATTR_START), GN_VISIBLE),
            generate_disasm_line(ref, 0));
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

static dump(f, ea, label, max_insns)
{
  dump_function(f, ea, label, max_insns);
  callers(f, ea, label);
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

static main()
{
  auto f;
  f = fopen("/host-tmp/moo2-late-oracle-ida-20260811.txt", "w");
  fprintf(f, "MOO2 LATE ORACLE STATIC IDA PROBE\n");
  fprintf(f, "tool=IDA Pro 9.4 idat IDC; static_only=true; runtime_not_executed\n");
  fprintf(f, "input_path=%s\n", get_input_file_path());
  fprintf(f, "input_sha256=recorded_by_container_command\n");
  fprintf(f, "address_basis=IDA linear addresses; raw names/operands preserved\n");
  fprintf(f, "min_ea=0x%X max_ea=0x%X\n", get_inf_attr(INF_MIN_EA), get_inf_attr(INF_MAX_EA));

  fprintf(f, "\n[ground_battle_and_rng]\n");
  dump(f, 0xEC15C, "raw_0xEC15C_Compute_Player_Ground_Combat_Bonuses", 900);
  dump(f, 0xEC3CE, "raw_0xEC3CE_Find_Troops", 1100);
  dump(f, 0xEC4FE, "raw_0xEC4FE_Compute_Ground_Combat_Info", 900);
  dump(f, 0xEC601, "raw_0xEC601_Ground_Combat_Round", 500);
  dump(f, 0xECE05, "raw_0xECE05_Change_Pop_Ownership", 1000);
  dump(f, 0xECECA, "raw_0xECECA_Resolve_Invasion_Troops", 1000);
  dump(f, 0xED260, "raw_0xED260_Change_Colony_Ownership", 1800);
  dump(f, 0xED59D, "raw_0xED59D_Invade", 700);
  dump(f, 0xED674, "raw_0xED674_Colony_Tank_Limit", 500);
  dump(f, 0xED713, "raw_0xED713_Unload_Transports", 1000);
  dump(f, 0xED7A3, "raw_0xED7A3_Compute_Colony_Ground_Combat_Info", 1000);
  dump(f, 0x124820, "raw_0x124820_Random", 400);
  dump(f, 0x12484C, "raw_0x12484C_Set_Random_Seed", 300);
  dump(f, 0x124878, "raw_0x124878_Get_Random_Seed", 300);
  dump(f, 0x79F26, "raw_0x79F26_Random_With_Restore", 500);

  fprintf(f, "\n[events_and_random_selection]\n");
  dump(f, 0x201F9, "raw_0x201F9_Init_Events", 500);
  dump(f, 0x2023F, "raw_0x2023F_Reset_Events", 500);
  dump(f, 0x2027E, "raw_0x2027E_Check_For_Event_Message", 700);
  dump(f, 0x2031D, "raw_0x2031D_Check_For_Event", 900);
  dump(f, 0x203CB, "raw_0x203CB_Start_Main_Event", 1200);
  dump(f, 0x20400, "raw_0x20400_Load_Event_Picture", 700);
  dump(f, 0x20460, "raw_0x20460_Drive_Event_Screen", 1200);
  dump(f, 0x20538, "raw_0x20538_Draw_Event_Screen", 1200);
  dump(f, 0x21B6D, "raw_0x21B6D_Setup_Next_Event", 900);
  dump(f, 0x2223D, "raw_0x2223D_Get_Event_Message", 1400);
  dump(f, 0x22D57, "raw_0x22D57_Determine_Event", 1400);
  dump(f, 0x22F5C, "raw_0x22F5C_Get_Event_Victim", 800);
  dump(f, 0x2310C, "raw_0x2310C_Player_Has_Bad_Event", 800);
  dump(f, 0x2325E, "raw_0x2325E_Event_Check_Industrial_Accident", 600);
  dump(f, 0x2332C, "raw_0x2332C_Event_Check_Mineral_Deposit", 600);
  dump(f, 0x2341E, "raw_0x2341E_Event_Check_Hyperspace_Flux", 700);
  dump(f, 0x2346E, "raw_0x2346E_Event_Check_Space_Anomaly", 700);
  dump(f, 0x234B8, "raw_0x234B8_Event_Check_Pirate_Activity", 700);
  dump(f, 0x23509, "raw_0x23509_Event_Check_Plague", 700);
  dump(f, 0x23536, "raw_0x23536_Event_Check_Warp_Beast", 700);
  dump(f, 0x23563, "raw_0x23563_Event_Check_Population_Boom", 700);
  dump(f, 0x2373F, "raw_0x2373F_Event_Check_Planet_Warning", 700);
  dump(f, 0x23780, "raw_0x23780_Event_Check_General_Status", 700);
  dump(f, 0x23B28, "raw_0x23B28_Pick_Completely_Random_Star", 500);
  dump(f, 0x23BEC, "raw_0x23BEC_Pick_Random_Player_In_Contact", 500);
  dump(f, 0x23CA1, "raw_0x23CA1_Pick_Random_Star", 800);
  dump(f, 0x23CED, "raw_0x23CED_Pick_Random_Officer", 600);
  dump(f, 0x23D44, "raw_0x23D44_Pick_Random_Ship", 600);
  dump(f, 0x23DA0, "raw_0x23DA0_Pick_Random_Colony_No_Outpost", 600);
  dump(f, 0x23DFE, "raw_0x23DFE_event_colony_filter", 600);
  dump(f, 0x245C4, "raw_0x245C4_Determine_Lucky_Players_Events", 900);
  dump(f, 0x21371, "raw_0x21371_Event_Twiddle", 2200);
  dump(f, 0x2230A, "raw_0x2230A_Event_Get_Fleet_Strength", 900);
  dump(f, 0x230B6, "raw_0x230B6_event_candidate_eligibility", 900);
  dump(f, 0x586D4, "raw_0x586D4_randomize_event_candidates", 800);
  dump(f, 0x10130A, "raw_0x10130A_Steal_App", 1200);
  dump(f, 0x1014A4, "raw_0x1014A4_N_Spies_Bonus", 1400);

  fprintf(f, "\n[damage_and_explosion_chain]\n");
  dump(f, 0x35821, "raw_0x35821_Apply_Internal_Damage", 700);
  dump(f, 0x3868F, "raw_0x3868F_Damage_From_Ship_Exploding", 1200);
  dump(f, 0x39DE0, "raw_0x39DE0_Apply_Ship_Damage", 1800);
  dump(f, 0x39F1D, "raw_0x39F1D_Destroy_Ship", 1200);
  dump(f, 0x3A6CA, "raw_0x3A6CA_Apply_Damage_To_Planet", 1300);
  dump(f, 0x40CC3, "raw_0x40CC3_Apply_Damage_To_Qship", 900);
  dump(f, 0x4257E, "raw_0x4257E_Resolve_Strat_Colony_Damage", 1600);
  dump(f, 0x3E563, "raw_0x3E563_Fighter_Combat_SFX", 900);
  dump(f, 0xA742B, "raw_0xA742B_Draw_Ship_Explosion", 700);
  dump(f, 0xAEC7B, "raw_0xAEC7B_Play_Ship_Explosion_SFX", 500);
  dump(f, 0x3897A, "raw_0x3897A_ship_explosion_eligibility", 900);
  dump(f, 0x39985, "raw_0x39985_apply_spherical_damage", 1400);
  dump(f, 0x40C2A, "raw_0x40C2A_explosion_damage_consumer", 900);
  dump(f, 0x494A8, "raw_0x494A8_engine_explosion_potential", 700);
  dump(f, 0x3A3C3, "raw_0x3A3C3_planet_damage_roll_consumer", 900);
  dump(f, 0x416CF, "raw_0x416CF_strat_damage_setup", 1000);
  dump(f, 0x41F80, "raw_0x41F80_strat_damage_step", 900);
  dump(f, 0x4221F, "raw_0x4221F_strat_damage_step", 900);
  dump(f, 0x420C0, "raw_0x420C0_strat_damage_step", 900);

  fprintf(f, "\n[cmbtshp_and_animation]\n");
  dump(f, 0x2A77D, "raw_0x2A77D_Combat_Ship_Class", 700);
  dump(f, 0x30631, "raw_0x30631_Draw_Ship", 1100);
  dump(f, 0x31F25, "raw_0x31F25_Combat_Draw_View_Ship", 1100);
  dump(f, 0x33199, "raw_0x33199_Get_Ship_Anim_Dimensions", 700);
  dump(f, 0x33CFA, "raw_0x33CFA_Display_Combat_View_Ship", 1100);
  dump(f, 0x30062, "raw_0x30062_Load_Combat_Ship_Sprite", 1500);
  dump(f, 0x49A41, "raw_0x49A41_Load_Combat_Ship", 1300);
  dump(f, 0x49F99, "raw_0x49F99_Load_Display_Ship_Sprite", 1200);
  dump(f, 0x5514C, "raw_0x5514C_Load_Antaran_Combat_Ship", 1100);
  dump(f, 0x55364, "raw_0x55364_Load_Small_Antaran_Combat_Ship", 900);
  dump(f, 0x5565D, "raw_0x5565D_Load_Medium_Antaran_Combat_Ship", 900);
  dump(f, 0x559DC, "raw_0x559DC_Load_Large_Antaran_Combat_Ship", 900);
  dump(f, 0x55E16, "raw_0x55E16_Load_Huge_Antaran_Combat_Ship", 900);
  dump(f, 0x562D6, "raw_0x562D6_Load_Titan_Antaran_Combat_Ship", 900);
  dump(f, 0x58697, "raw_0x58697_Get_Ship_Id_Picture_Seg", 700);
  dump(f, 0x5869B, "raw_0x5869B_Get_Ship_Picture_Seg", 700);
  dump(f, 0x586D3, "raw_0x586D3_Get_Ship_Picture_Segk", 700);
  dump(f, 0x3F5F1, "raw_0x3F5F1_Move_Ship_Heading", 1200);
  dump(f, 0x3F628, "raw_0x3F628_Get_Facing_Heading", 900);

  fprintf(f, "\n[diplomacy_trade_and_gifts]\n");
  dump(f, 0x5232E, "raw_0x5232E_Start_Trade_Treaty", 1000);
  dump(f, 0x533F4, "raw_0x533F4_Diplomacy_Test", 1800);
  dump(f, 0x539D9, "raw_0x539D9_Get_Gift_Response", 1100);
  dump(f, 0x53EDB, "raw_0x53EDB_Get_Player_Diplomacy_Personality", 600);
  dump(f, 0x101BA4, "raw_0x101BA4_Base_Trade_Agreement_Goal", 900);
  dump(f, 0x101B3C, "raw_0x101B3C_Trade_Agreement_Base", 900);
  dump(f, 0x101C93, "raw_0x101C93_Trade_Agreement_Response", 1000);
  dump(f, 0x101CC5, "raw_0x101CC5_Research_Agreement_Response", 1000);
  dump(f, 0x101E77, "raw_0x101E77_Process_Trade_And_Research_Agreements", 1800);
  dump(f, 0x101EE3, "raw_0x101EE3_Start_Trade_Agreement", 1800);
  dump(f, 0x101F82, "raw_0x101F82_Start_Research_Agreement", 1000);
  dump(f, 0xDCC83, "raw_0xDCC83_AI_Agrees_To_Trade_Agreement", 1200);
  dump(f, 0x524C3, "raw_0x524C3_Advance_Trade_Value", 1000);
  dump(f, 0x524FB, "raw_0x524FB_Trade_Relation_Delta", 1000);

  fprintf(f, "\n[spy_storage_and_bonuses]\n");
  dump(f, 0x100BC5, "raw_0x100BC5_Compute_Spy_Bonuses", 1400);
  dump(f, 0x1026F1, "raw_0x1026F1_Get_Their_Spy_Number", 500);
  dump(f, 0x102711, "raw_0x102711_Get_Their_Spy_Mission", 500);
  dump(f, 0x102739, "raw_0x102739_Get_My_Spy_Number", 500);
  dump(f, 0x10275F, "raw_0x10275F_Get_My_Spy_Mission", 500);
  dump(f, 0x102776, "raw_0x102776_Get_My_Agent_Number", 500);
  dump(f, 0x10278D, "raw_0x10278D_Get_My_Agent_Mission", 500);
  dump(f, 0x1027B5, "raw_0x1027B5_Spy_Storage_Setter", 600);
  dump(f, 0x10282D, "raw_0x10282D_Spy_Storage_Setter", 600);
  dump(f, 0x10289D, "raw_0x10289D_Spy_Storage_Setter", 600);
  dump(f, 0x10290D, "raw_0x10290D_Spy_Storage_Setter", 600);
  dump(f, 0x10297D, "raw_0x10297D_Spy_Storage_Setter", 600);
  dump(f, 0x1029D1, "raw_0x1029D1_Spy_Storage_Setter", 600);

  fprintf(f, "\n[leaders_and_eta]\n");
  dump(f, 0x1307F, "raw_0x1307F_Init_Officers", 700);
  dump(f, 0x934CF, "raw_0x934CF_Deassign_Officer", 900);
  dump(f, 0x933F2, "raw_0x933F2_Check_Officer_Fields", 900);
  dump(f, 0x93528, "raw_0x93528_Decrement_Officer_ETA", 700);
  dump(f, 0x98F42, "raw_0x98F42_Get_Ship_Leader_ETA", 1100);
  dump(f, 0x94951, "raw_0x94951_Get_Leader_Experience_Level", 700);
  dump(f, 0x93D4B, "raw_0x93D4B_Leader_Experience_Bucket", 700);
  dump(f, 0x943A0, "raw_0x943A0_Move_From_Limbo_To_Pool", 900);
  dump(f, 0x9453C, "raw_0x9453C_Move_Officer_To_Pool", 900);
  dump(f, 0x9467D, "raw_0x9467D_Officer_Exp", 700);
  dump(f, 0x957E3, "raw_0x957E3_Print_ETA_On_Officer_Picture", 700);
  dump(f, 0x97287, "raw_0x97287_Set_Officer_To_Player", 1000);
  dump(f, 0x9776C, "raw_0x9776C_Officer_Status_Ok", 700);
  dump(f, 0x979A0, "raw_0x979A0_Officer_Experience_Adjustment", 900);
  dump(f, 0x97A66, "raw_0x97A66_Random_Officer_Check", 1200);
  dump(f, 0x97AD4, "raw_0x97AD4_Officer_Offer_Writeback", 1200);
  dump(f, 0x97B2D, "raw_0x97B2D_Generate_Random_Officer", 1400);
  dump(f, 0x97C64, "raw_0x97C64_Add_Officer_Experience", 1000);
  dump(f, 0xD7439, "raw_0xD7439_Do_AI_Leaders", 1800);
  dump(f, 0xDCDAC, "raw_0xDCDAC_Handle_Leader_At_Lost_Colony", 900);
  dump(f, 0xE1FC7, "raw_0xE1FC7_Leader_At_Anomaly", 900);

  data_refs(f, 0x1829F1, "raw_global_0x1829F1_CMBTSFX_LBX");
  raw_bytes(f, 0x1829F1, 0x20, "raw_global_0x1829F1_CMBTSFX_LBX_slice");
  raw_bytes(f, 0x17F642, 0x90, "raw_table_0x17F642_ship_class_power_candidate");
  raw_bytes(f, 0x17FD15, 0x80, "raw_table_0x17FD15_weapon_modifier_candidate");
  raw_bytes(f, 0x18105C, 0x20, "raw_table_0x18105C_trade_agreement_values");
  raw_bytes(f, 0x180DB8, 0x20, "raw_table_0x180DB8_diplomacy_personality_values");
  raw_bytes(f, 0x180CCC, 0x20, "raw_table_0x180CCC_gift_response_values");
  raw_bytes(f, 0x181070, 0x20, "raw_table_0x181070_trade_event_values");
  raw_bytes(f, 0x17EB3D, 0x13 * 49, "raw_table_0x17EB3D_building_records");
  data_refs(f, 0x17F642, "raw_table_0x17F642_ship_class_power_candidate");
  data_refs(f, 0x17FD15, "raw_table_0x17FD15_weapon_modifier_candidate");

  fclose(f);
  qexit(0);
}
