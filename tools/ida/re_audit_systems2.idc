// MOO2 忠實度第二輪稽核：研究、維護、事件、殖民、星際移動、間諜與安塔蘭。
// 僅在由 Orion2.exe 建立的一次性資料庫執行；不改名、不套型別、不保存資料庫。

#include <idc.idc>

static dump_callers(f, ea)
{
  auto ref;
  auto count;
  ref = get_first_cref_to(ea);
  count = 0;
  while (ref != BADADDR && count < 120)
  {
    fprintf(f, "  caller site=0x%X function=0x%X raw_name=%s insn=%s\n",
            ref, get_func_attr(ref, FUNCATTR_START),
            get_func_name(get_func_attr(ref, FUNCATTR_START)),
            generate_disasm_line(ref, 0));
    ref = get_next_cref_to(ea, ref);
    count++;
  }
  fprintf(f, "  caller_count_reported=%d\n", count);
}

static dump_call_targets_from_site(f, site)
{
  auto target;
  auto count;
  target = get_first_cref_from(site);
  count = 0;
  while (target != BADADDR && count < 16)
  {
    if (get_func_attr(target, FUNCATTR_START) == target || get_name(target, GN_VISIBLE) != "")
      fprintf(f, "    call_xref_target=0x%X raw_name=%s\n",
              target, get_name(target, GN_VISIBLE));
    target = get_next_cref_from(site, target);
    count++;
  }
}

static dump_function(f, ea, ledger_label, max_insns)
{
  auto start;
  auto end;
  auto cur;
  auto count;
  auto op;
  auto target;
  auto mnem;
  start = get_func_attr(ea, FUNCATTR_START);
  end = get_func_attr(ea, FUNCATTR_END);
  fprintf(f, "\n[function]\nledger_label=%s requested=0x%X raw_name=%s start=0x%X end=0x%X\n",
          ledger_label, ea, get_name(ea, GN_VISIBLE), start, end);
  if (start == BADADDR || end == BADADDR)
  {
    fprintf(f, "status=no_function\n");
    return;
  }
  dump_callers(f, start);
  cur = start;
  count = 0;
  while (cur != BADADDR && cur < end && count < max_insns)
  {
    fprintf(f, "  0x%X %s\n", cur, generate_disasm_line(cur, 0));
    mnem = print_insn_mnem(cur);
    if (mnem == "call" || mnem == "callf")
    {
      target = get_operand_value(cur, 0);
      fprintf(f, "    raw_operand_target=0x%X raw_name=%s\n",
              target, get_name(target, GN_VISIBLE));
      dump_call_targets_from_site(f, cur);
    }
    for (op = 0; op < 8; op++)
    {
      if (get_operand_type(cur, op) == o_mem || get_operand_type(cur, op) == o_displ)
      {
        target = get_operand_value(cur, op);
        if (is_loaded(target) && !is_code(get_full_flags(target)))
          fprintf(f, "    data operand=%d target=0x%X raw_name=%s\n",
                  op, target, get_name(target, GN_VISIBLE));
      }
    }
    cur = next_head(cur, end);
    count++;
  }
  if (cur != BADADDR && cur < end)
    fprintf(f, "  status=instruction_limit limit=%d\n", max_insns);
}

static main()
{
  auto f;
  Wait();
  f = fopen("/audit/ida-systems2-audit.txt", "w");
  fprintf(f, "MOO2 GAMEPLAY SYSTEMS SECOND IDA RE-AUDIT\n");
  fprintf(f, "tool=IDA Pro 9.4 idat IDC; static_only=true\n");
  fprintf(f, "input=%s\naddress_basis=IDA linear\n", get_input_file_path());

  dump_function(f, 0x136B3, "Next_Turn_Calc_", 1300);
  dump_function(f, 0xDFF74, "Colony_Research_Production_", 900);
  dump_function(f, 0xE1EF4, "Chance_For_Research_Breakthrough_", 500);
  dump_function(f, 0xE44E0, "Check_For_Research_Breakthrough_", 900);
  dump_function(f, 0xE2000, "Compute_Player_Maintenance_", 900);
  dump_function(f, 0xEE0B0, "Player_Maintenance_", 900);
  dump_function(f, 0x21371, "Setup_Next_Event_", 1000);
  dump_function(f, 0x2230A, "Determine_Event_", 1400);
  dump_function(f, 0xBB082, "Colonize_Planet_", 1100);
  dump_function(f, 0xFFEEA, "Move_All_Ships_Toward_Stars_", 1100);
  dump_function(f, 0x100A83, "Compute_Spy_Bonuses_", 700);
  dump_function(f, 0x1028D5, "Take_My_Spy_", 500);
  dump_function(f, 0x102982, "Add_My_Spy_", 500);
  dump_function(f, 0x63D92, "Antaran_Invasion_Check_", 700);

  fclose(f);
  qexit(0);
}
