/*
** ret.c
** request Return
** inherits from ssl
*/

#include <stdlib.h>
#include <string.h>

#include "ssl.h"
#include "ret.h"

struct _RET
{
  int code;  // Code de retour ou d'erreur
  SSL ssl;   // Objet Simple String List
};

RET RET_new(unsigned int size, int code)
{
  RET ret;

  ret = (RET)malloc(sizeof(struct _RET));

  if ( ret != NULL )
  {
    ret->code = code;
    if ( ( ret->ssl = SSL_new(size) ) == NULL ) return RET_del(ret);
  }

  return ret;
}

RET RET_del(RET ret)
{
  if ( ret != NULL )
  {
    SSL_del(ret->ssl);
    free(ret);
  }

  return NULL;
}

int RET_code(RET ret)
{
  if ( ret != NULL )
    return ret->code;
  else
    return 0;
}

static SSL RET_getSSL(RET ret)
{
  if ( ret != NULL )
    return ret->ssl;
  else
    return NULL;
}

unsigned int RET_count(RET ret)
{
  return SSL_count(RET_getSSL(ret));
}

unsigned int RET_add(RET ret, char *sz) // Zero terminated String (0t)
{
  return SSL_add2(RET_getSSL(ret), sz, strlen(sz));
}

unsigned int RET_add2(RET ret, char *str, unsigned int len)
{
  return SSL_add2(RET_getSSL(ret), str, len);
}

char *RET_get(RET ret, int index)
{
  return SSL_get(RET_getSSL(ret), index);
}

unsigned int RET_len(RET ret, int index)
{
  return SSL_len(RET_getSSL(ret), index);
}

/* eof */