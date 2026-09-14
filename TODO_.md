# TODO:

- [Opsional] Performance reporter, berapa lama suatu fungsi yang gw taro itu dieksekusi
- [Opsional] Ganti format nama logger jadi tanggal, jangan pake unix time, biar file ga numpuk

## **Basic: Target Rabu selesai (Kelarnya malah Kamis)**

- [x] ~~Subject diff generator (add, del, changed)~~
- [x] ~~Class diff generator (add, del, changed)~~
- [x] ~~Schedule diff generator (add, del, changed)~~
- [x] ~~Apply all of these diffs, oh my god bruhhh~~
- [x] ~~Test all of above, but idk how (Partially tested)~~
- [x] ~~Bikin cronjob perjam buat nge fetch data jadwal~~

## **Advance: Target MAKSIMAL Sabtu selesai**

- [x] ~~Bikin tabel class follower, buat dapet notif perubahan matkul/kelas/jadwal, sama buat pengingat presensi (perhaps, make pg_ivm?)~~
- [ ] Modifikasi fungsi cronjob di "Basic" tadi buat manggil fungsi diff string generator dan dikirim ke class follower (ya ya ya, fuck)
- [ ] Bikin cronjob perhari buat ambil siapa yang perlu dapet pengingat presensi pada hari itu (should be easy, hm, hm, ya)
- [ ] Open test selama sepekan lebih dikit (memastikan kestabilannya, paling tidak sampai Minggu, 15 Februari 2025 -> pekan pertama kuliah)
- [x] ~~Bikin command biar pengguna bisa nge follow kelas~~
- [x] ~~Bikin command biar pengguna bisa ngasih offset waktu pengingat, offset bisa sebelum ataupun sesudahx~~

## **Next: Target Akhir Februari**

- [ ] Balikin semua command yang udah pernah dibikin, di-copy aja dari base lama, ga usah dipikir lagi aowkoakwa
- [ ] Tes stabilitas semua command

## **Keluh Kesah**

- Duh males banget dah mainan database, gatau kenapa, males banget
