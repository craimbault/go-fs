// maingofs.c
#include <stdlib.h>
#include <stdio.h>

#include "libgofs.h"
#include "ret.h"

int main() {
  RET ret = NULL;

  gofs_write("filetest1.txt", "Contenu de test 1\r\nEnd");
  gofs_write("filetest2.txt", "Contenu de test 2\r\nEnd");

  ret = gofs_stat("filetest1.txt");
  printf(">  stat return with code : %d\n", RET_code(ret));
  for ( int i = 0; i < RET_count(ret); i++ )
    printf(">%3d: %s\n", i+1, RET_get(ret, i));
  ret = RET_del(ret);

  ret = gofs_list("");
  printf(">  list return with code : %d\n", RET_code(ret));
  for ( int i = 0; i < RET_count(ret); i++ )
    printf(">%3d: %s\n", i+1, RET_get(ret, i));
  ret = RET_del(ret);

  ret = gofs_read("filetest1.txt");
  printf(">  read return with code : %d\n", RET_code(ret));
  printf(">  contenu du fichier lu :\n%.*s\n", RET_len(ret, 0), RET_get(ret, 0));
  ret = RET_del(ret);

  gofs_move("filetest1.txt", "moved/filetest.txt");

  ret = gofs_list("moved/");
  printf(">  list on moved/ return with code : %d\n", RET_code(ret));
  for ( int i = 0; i < RET_count(ret); i++ )
    printf(">%3d: %s\n", i+1, RET_get(ret, i));
  ret = RET_del(ret);

  gofs_delete("moved/filetest.txt");

  ret = gofs_list("");
  printf(">  list return with code : %d\n", RET_code(ret));
  for ( int i = 0; i < RET_count(ret); i++ )
    printf(">%3d: %s\n", i+1, RET_get(ret, i));
  ret = RET_del(ret);

  return 0;
}