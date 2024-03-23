/*
** ret.h
** request Return
** inherits from ssl
*/

#ifndef RET_H  /* Pour éviter les déclarations multiples */
#define RET_H

typedef struct _RET *RET;

RET RET_new(unsigned int size, int code);
RET RET_del(RET ssl);
int RET_code(RET ret);
unsigned int RET_count(RET ret);
unsigned int RET_add(RET ret, char *sz);
unsigned int RET_add2(RET ret, char *str, unsigned int len);
char *RET_get(RET ret, int index);
unsigned int RET_len(RET ret, int index);

#endif  /* RET_H */

/* eof */