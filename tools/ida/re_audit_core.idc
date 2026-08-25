// MOO2 忠實度重新稽核：五類高影響玩家機制的原始 IDA 函式、caller 與 data xref。
// 在由 Orion2.exe 新建的一次性資料庫執行；保留原始位址、名稱與運算元，不修改資料庫。

#include <idc.idc>

static disasm_window(f, ea, before, after)
{
  auto cur;
  auto start;
  auto end;
  auto i;
  start = get_func_attr(ea, FUNCATTR_START);
  end = get_func_attr(ea, FUNCATTR_END);
  cur = ea;
  i = 0;
  while (i < before && cur != BADADDR && cur > start)
  {
    cur = prev_head(cur, start);
    i++;
  }
  i = 0;
  while (cur != BADADDR && cur < end && i < before + after + 1)
  {
    fprintf(f, "    0x%X %s\n", cur, generate_disasm_line(cur, 0));
    cur = next_head(cur, end);
    i++;
  }
}

static callers(f, ea)
{
  auto ref;
  auto count;
  ref = get_first_cref_to(ea);
  count = 0;
  while (ref != BADADDR && count < 80)
  {
    fprintf(f, "  caller=0x%X caller_func=0x%X caller_name=%s\n",
            ref, get_func_attr(ref, FUNCATTR_START),
            get_func_name(get_func_attr(ref, FUNCATTR_START)));
    disasm_window(f, ref, 8, 5);
    ref = get_next_cref_to(ea, ref);
    count++;
  }
  fprintf(f, "  caller_count_reported=%d\n", count);
}

static data_refs_from_function(f, start, end)
{
  auto cur;
  auto op;
  auto target;
  auto flags;
  auto seen;
  seen = 0;
  cur = start;
  while (cur != BADADDR && cur < end)
  {
    for (op = 0; op < 8; op++)
    {
      if (get_operand_type(cur, op) == o_mem || get_operand_type(cur, op) == o_displ)
      {
        target = get_operand_value(cur, op);
        flags = get_full_flags(target);
        if (is_loaded(target) && !is_code(flags))
        {
          fprintf(f, "  data_operand site=0x%X operand=%d target=0x%X target_name=%s insn=%s\n",
                  cur, op, target, get_name(target, GN_VISIBLE), generate_disasm_line(cur, 0));
          seen++;
        }
      }
    }
    cur = next_head(cur, end);
  }
  fprintf(f, "  data_operands_reported=%d\n", seen);
}

static dump_call_targets_from_site(f, site)
{
  auto target;
  auto count;
  target = get_first_cref_from(site);
  count = 0;
  while (target != BADADDR && count < 16)
  {
    // IDA 的 code xref 可能同時含 fall-through；只保留函式入口或具名 code target。
    if (get_func_attr(target, FUNCATTR_START) == target || get_name(target, GN_VISIBLE) != "")
      fprintf(f, "    call_xref_target=0x%X target_name=%s\n",
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
  auto mnem;
  auto target;
  start = get_func_attr(ea, FUNCATTR_START);
  end = get_func_attr(ea, FUNCATTR_END);
  fprintf(f, "\n[function]\nledger_label=%s requested=0x%X raw_name=%s start=0x%X end=0x%X\n",
          ledger_label, ea, get_name(ea, GN_VISIBLE), start, end);
  if (start == BADADDR || end == BADADDR)
  {
    fprintf(f, "status=no_function\n");
    return;
  }
  callers(f, start);
  data_refs_from_function(f, start, end);
  cur = start;
  count = 0;
  while (cur != BADADDR && cur < end && count < max_insns)
  {
    fprintf(f, "  0x%X %s\n", cur, generate_disasm_line(cur, 0));
    mnem = print_insn_mnem(cur);
    if (mnem == "call" || mnem == "callf")
    {
      target = get_operand_value(cur, 0);
      fprintf(f, "    raw_operand_target=0x%X target_name=%s\n", target, get_name(target, GN_VISIBLE));
      dump_call_targets_from_site(f, cur);
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
  f = fopen("/audit/ida-core-audit.txt", "w");
  fprintf(f, "MOO2 CORE MECHANICS IDA RE-AUDIT\n");
  fprintf(f, "tool=IDA Pro 9.4 idat IDC; static_only=true\n");
  fprintf(f, "input=%s\naddress_basis=IDA linear\n", get_input_file_path());

  dump_function(f, 0x136B3, "Next_Turn_Calc_", 500);
  dump_function(f, 0xE2DCA, "Apply_Colony_Pop_Growth_", 600);
  dump_function(f, 0x53146, "Diplomacy_Test_", 900);
  dump_function(f, 0x2552D, "NPC_To_NPC_Treaty_Negotiations_", 900);
  dump_function(f, 0x4E3B5, "Change_Relations_", 600);
  dump_function(f, 0xEC4FE, "Ground_Combat_Round_", 700);
  dump_function(f, 0xEC601, "Resolve_Ground_Combat_", 900);
  dump_function(f, 0x4257E, "Strategic_Bombardment_", 700);
  dump_function(f, 0x15B90, "Calc_Council_Vote_", 500);
  dump_function(f, 0x15EBC, "Council_Votes_", 800);
  dump_function(f, 0x168AF, "Check_For_Council_Meeting_", 600);
  dump_function(f, 0x3AC20, "Fire_Fighter_Bomb_", 500);
  dump_function(f, 0x3AD57, "Fire_Fighter_Beam_", 700);
  dump_function(f, 0x5F64C, "Fighter_Garrison_Strength_", 600);
  dump_function(f, 0x42371, "Get_Colony_Hits_", 600);

  fclose(f);
  qexit(0);
}
