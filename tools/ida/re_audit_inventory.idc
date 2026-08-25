// MOO2 忠實度重新稽核：在一次性 IDA 資料庫輸出原始函式邊界與資料庫覆蓋資訊。
// 不改名、不套型別、不保存來源資料庫；輸出位置由固定的 /audit 掛載提供。

#include <idc.idc>

static main()
{
  auto f;
  auto ea;
  auto functions;
  auto named_functions;
  auto code_bytes;
  auto data_bytes;

  Wait();
  f = fopen("/audit/ida-inventory.txt", "w");
  fprintf(f, "MOO2 IDA RE-AUDIT INVENTORY\n");
  fprintf(f, "tool=IDA Pro 9.4 idat IDC\n");
  fprintf(f, "input=%s\n", get_input_file_path());
  fprintf(f, "address_basis=IDA linear\n");
  fprintf(f, "min_ea=0x%X max_ea=0x%X\n", get_inf_attr(INF_MIN_EA), get_inf_attr(INF_MAX_EA));

  functions = 0;
  named_functions = 0;
  ea = get_next_func(0);
  while (ea != BADADDR)
  {
    functions++;
    if (substr(get_func_name(ea), 0, 4) != "sub_")
      named_functions++;
    ea = get_next_func(ea);
  }

  code_bytes = 0;
  data_bytes = 0;
  ea = get_inf_attr(INF_MIN_EA);
  while (ea != BADADDR && ea < get_inf_attr(INF_MAX_EA))
  {
    if (is_code(get_full_flags(ea)))
      code_bytes++;
    else if (is_data(get_full_flags(ea)))
      data_bytes++;
    ea++;
  }
  fprintf(f, "functions=%d named_functions=%d code_bytes=%d data_bytes=%d\n",
          functions, named_functions, code_bytes, data_bytes);
  fclose(f);
  qexit(0);
}
